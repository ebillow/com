package snet

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"server/com/util"
	"server/pb"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"server/com/log"
)

var (
	errAPINotFind   = errors.New("api not defined")
	errMsgParse     = errors.New("parser msg error")
	errServerClosed = errors.New("server closed")
	errMataDataErr  = errors.New("mata data err")
)

var ctrl = make(chan struct{})
var waitClose = make(chan bool)

// Server rpc service
type Server struct {
	pb.UnimplementedSrvServiceServer
}

// SrvSrv rpc stream
func (server *Server) SrvSrv(stream pb.SrvService_SrvSrvServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		log.Error("cannot read metadata from context")
		return errMataDataErr
	}

	if len(md["id"]) == 0 {
		log.Error("cannot read key:id from metadata")
		return errMataDataErr
	}
	if len(md["connector"]) == 0 {
		log.Error("cannot read key:connector from metadata")
		return errMataDataErr
	}
	sess := NewSession(md["connector"][0], util.ParseUint32(md["id"][0]), stream)

	go recvLoop(sess, stream)
	go sendLoop(sess, stream, sess.desc, sess.die)

	//log.Debugf("start stream success:%s", sess.desc)

	return handLoop(sess)
}

// StartServe 开始服务监听
func StartServe(ctx context.Context, wait *sync.WaitGroup, ip string, port uint16) error {
	addr := ":" + strconv.Itoa(int(port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panic(err)
		return err
	}

	s := grpc.NewServer()
	ins := &Server{}
	pb.RegisterSrvServiceServer(s, ins)

	wait.Add(1)
	go func() {
		select {
		case <-ctx.Done():
			log.Infof("snet start stop")
			waitMonitorReturn()
			s.GracefulStop()
			log.Infof("snet grpc graceful stop")
			wait.Done()
		}
	}()

	RegisterService(localName, localId, ip, port)
	atomic.StoreUint32(&netStart, 1)

	log.Infof("start serve. tcp on:%v ", listener.Addr())
	log.Infof(util.SuccessShow)

	return s.Serve(listener)
}

func waitMonitorReturn() {
	defer close(waitClose)

	unRegisterService(localName, localId)
	t := time.NewTimer(time.Second * 5)
	for {
		select {
		case waitClose <- true:
			return
		case <-t.C:
			return
		}
	}
}

func Close() {
	close(ctrl)
	log.Trace("snet close ctrl")
}

func unRegisterService(svcType string, id uint32) {
	s := GetMonitorSession()
	if s == nil {
		log.Warnf("can not get monitor session")
	} else {
		s.Send(pb.MsgIDS2S_SvcService, &pb.MsgServiceInfo{
			Op:          pb.MsgServiceInfo_OpDel,
			Name:        svcType,
			Id:          id,
			Time:        time.Now().Unix(),
			IsToMonitor: true,
		})
		log.Infof("del service %s %d", svcType, id)
		return
	}
}

func RegisterService(svcType string, id uint32, listenIP string, port uint16) {
	register := func(now time.Time) {
		s := GetMonitorSession()
		if s == nil {
			log.Warnf("can not get monitor session")
		} else {
			msg := &pb.MsgServiceInfo{
				Op:          pb.MsgServiceInfo_OpUpdate,
				Name:        svcType,
				Id:          id,
				EndPoint:    listenIP + ":" + util.ToString(port),
				Time:        now.Unix(),
				IsToMonitor: true,
			}
			s.Send(pb.MsgIDS2S_SvcService, msg)
			//log.Infof("update service %s %d, %s", svcType, id, listenIP+":"+util.ToString(port))
		}
	}
	go func() {
		tLoop := time.NewTicker(time.Second * 3)
		defer func() {
			tLoop.Stop()
		}()
		register(time.Now())
		for {
			select {
			case now := <-tLoop.C:
				register(now)
			case <-ctrl:
				return
			}
		}
	}()
}
