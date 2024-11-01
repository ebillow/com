package snet

import (
	"context"
	"google.golang.org/grpc/metadata"
	"server/com/log"
	"server/com/snet/service"
	"server/com/util"
	"server/pb"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	target    string
	targetID  uint32
	localName string
	localID   uint32
	callback  func(svcId uint32)

	timeOutCnt int32
	heartbeat  int64
}

var (
	connOnce       sync.Once
	conns          sync.Map
	svcMgrType     = "svcMgrForSNET"
	heartBeat      = int64(20)
	timeOutCntMax  = int32(3)
	startReconnect = true
)

const (
	heartBeatCheckSpan = 10
)

func Init(monitorEndPoint string, selfName string, selfId uint32) {
	connMonitor(monitorEndPoint, selfName, selfId)
}

func DelayReconnect() {
	startReconnect = false
}

func StartReconnect(span int64) {
	if span != 0 {
		heartBeat = span
		if heartBeat < heartBeatCheckSpan+5 {
			heartBeat = heartBeatCheckSpan + 5
		}
		log.Infof("set heart beat time=%d", heartBeat)
	}
	startReconnect = true
}

func Connect(targetName string, connCallback func(svcId uint32), wait chan uint32) {
	connOnce.Do(func() {
		go checkReconnect()
	})

	service.RegisterDialSvcCallback(targetName, func(svcId uint32) {
		atomic.StoreUint32(&netStart, 1)

		conn := &Conn{
			heartbeat: time.Now().Unix(),
			target:    targetName,
			targetID:  svcId,
			localName: localName,
			localID:   localId,
			callback:  connCallback,
		}
		conns.Store(makeSesDesc(targetName, svcId), conn)
		log.Infof("register %s %d connect callback", targetName, svcId)

		//log.Debugf("%s %d callback", targetName, svcId)
		s := StartStream(targetName, svcId, localName, localId)
		if s == nil {
			log.Errorf("start stream to %s%d err", targetName, svcId)
			return
		}
	}, wait)

	service.NeedService(targetName)
}

func GetCallback(svcType string, id uint32) func(svcId uint32) {
	conn, ok := conns.Load(makeSesDesc(svcType, id))
	if ok && conn.(*Conn) != nil {
		return conn.(*Conn).callback
	}
	return nil
}

func checkReconnect() {
	tLoop := time.NewTicker(time.Second * heartBeatCheckSpan)
	defer tLoop.Stop()

	for {
		select {
		case now := <-tLoop.C:
			if !startReconnect {
				continue
			}
			conns.Range(func(key, value interface{}) bool {
				conn := value.(*Conn)
				s := GetSession(conn.target, conn.targetID)
				if s == nil {
					log.Infof("%s %d try reconnect, because s == nil", conn.target, conn.targetID)
					StartStream(conn.target, conn.targetID, conn.localName, conn.localID)
					atomic.AddInt64(&conn.heartbeat, now.Unix())
				} else {
					if now.Unix()-atomic.LoadInt64(&conn.heartbeat) > heartBeat {
						timeOutCnt := atomic.AddInt32(&conn.timeOutCnt, 1)
						log.Infof("%s %d time out:%d", conn.target, conn.targetID, timeOutCnt)
						if timeOutCnt >= timeOutCntMax {
							log.Infof("%s %d try reconnect, because heartbeat > %d", conn.target, conn.targetID, heartBeat)
							s.CloseSync()

							s = StartStream(conn.target, conn.targetID, conn.localName, conn.localID)
							atomic.StoreInt64(&conn.heartbeat, now.Unix())
							atomic.StoreInt32(&conn.timeOutCnt, 0)
						}
					}
				}
				if s != nil {
					s.Send(pb.MsgIDS2S_SvcHeartbeat, nil)
				}
				return true
			})

		case <-ctrl:
			return
		}
	}
}

// StartStream 开启到game的流
func StartStream(targetName string, targetID uint32, localName string, localId uint32) *Session {
	conn := service.Get(targetName, targetID)
	if conn == nil {
		log.Errorf("can not get service %s_%d", targetName, targetID)
		return nil
	}
	cli := pb.NewSrvServiceClient(conn)
	mtdata := metadata.New(map[string]string{"connector": localName, "id": util.ToString(localId)})
	ctx := metadata.NewOutgoingContext(context.Background(), mtdata)

	stream, err := cli.SrvSrv(ctx)
	if err != nil {
		log.Errorf("start stream %s %d %s err:%v", targetName, targetID, conn.Target(), err)
		return nil
	}

	s := NewSession(targetName, targetID, stream)

	go recvLoop(s, stream)
	go sendLoop(s, stream, s.desc, s.die)
	go func() {
		_ = handLoop(s)
		_ = stream.CloseSend()
	}()
	s.wait.Wait()

	return s
}
