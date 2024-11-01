package snet

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"net"
	"server/com/log"
	"server/com/util"
	"server/pb"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	allSvc     = make(map[string]*pb.MsgServer)
	update     = make(map[string]*pb.MsgServer)
	mtxMonitor sync.Mutex

	localName string
	localId   uint32

	monitorEndPoint string
	monitorOnce     sync.Once
)

func OpenMonitor(port uint16) error {
	addr := ":" + strconv.Itoa(int(port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panic(err)
		return err
	}

	s := grpc.NewServer()
	ins := &Server{}
	pb.RegisterSrvServiceServer(s, ins)

	go func() {
		t := time.NewTicker(time.Second)
		defer func() {
			log.Info("monitor stop")
			s.GracefulStop()
			t.Stop()
		}()

		for {
			select {
			case <-ctrl:
				return
			case <-t.C:
				updateServers()
			}
		}
	}()

	atomic.StoreUint32(&netStart, 1)
	log.Infof("start monitor. tcp on:%v ", listener.Addr())

	return s.Serve(listener)
}

func updateServers() {
	msg := pb.MsgServiceInfo{Op: pb.MsgServiceInfo_OpUpdate}

	mtxMonitor.Lock()
	if len(update) == 0 {
		mtxMonitor.Unlock()
		return
	}
	msg.All = update

	b, err := proto.Marshal(&msg)
	if err != nil {
		log.Warnf("marshal err %v", err)
		return
	}
	update = make(map[string]*pb.MsgServer)
	mtxMonitor.Unlock()

	sess := AllSessions()
	for _, s := range sess {
		s.SendByte(uint32(pb.MsgIDS2S_SvcService), b)
	}
}

func makeKey(name string, id uint32) string {
	return name + "_" + util.Uint32ToString(id)
}

func OnMonitorProto(msg *pb.MsgServiceInfo, s *Session) {
	switch msg.Op {
	case pb.MsgServiceInfo_OpReqAll:
		mtxMonitor.Lock()
		msg.All = allSvc
		msg.IsToMonitor = false
		s := GetSession(msg.Name, msg.Id)
		if s != nil {
			s.Send(pb.MsgIDS2S_SvcService, msg)
		}
		mtxMonitor.Unlock()
	case pb.MsgServiceInfo_OpUpdate:
		//if msg.Name == share.AccountTopic && msg.Id == 22 {
		//	log.Infof("%s", msg.String())
		//}
		key := makeKey(msg.Name, msg.Id)
		mtxMonitor.Lock()
		ser := &pb.MsgServer{
			Type:     msg.Name,
			Id:       msg.Id,
			EndPoint: msg.EndPoint,
		}
		allSvc[key] = ser
		update[key] = ser
		mtxMonitor.Unlock()

	case pb.MsgServiceInfo_OpDel:
		key := makeKey(msg.Name, msg.Id)
		mtxMonitor.Lock()
		delete(allSvc, key)
		delete(update, key) //处理时序问题

		mtxMonitor.Unlock()
		s := GetSession(msg.Name, msg.Id)
		if s != nil {
			s.Send(pb.MsgIDS2S_SvcService, &pb.MsgServiceInfo{Op: pb.MsgServiceInfo_OpDelRet, Name: msg.Name, Id: msg.Id})
		}
		sess := AllSessions()
		msg.IsToMonitor = false

		var b []byte
		var err error

		if msg != nil {
			b, err = proto.Marshal(msg)
			if err != nil {
				log.Warnf("%s marshal err %v when send:%s", s.desc, err, msg.String())
				return
			}
		}

		for _, s := range sess {
			s.SendByte(uint32(pb.MsgIDS2S_SvcService), b)
		}
	}
}
