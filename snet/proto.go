package snet

import (
	"google.golang.org/protobuf/proto"
	"server/com/snet/service"
	"server/pb"
	"sync/atomic"
	"time"
)

func init() {
	RegisterHandle(pb.MsgIDS2S_SvcHeartbeatRet, func() proto.Message { return &pb.MsgServiceInfo{} }, onHeartbeatRet)
	RegisterHandle(pb.MsgIDS2S_SvcService, func() proto.Message { return &pb.MsgServiceInfo{} }, onRecvServiceInfo)
	RegisterHandle(pb.MsgIDS2S_SvcHeartbeat, func() proto.Message { return nil }, onHeartbeat)
}

func onRecvServiceInfo(msgBase proto.Message, s *Session) {
	msg, ok := msgBase.(*pb.MsgServiceInfo)
	if !ok || msg == nil {
		return
	}
	if msg.IsToMonitor {
		OnMonitorProto(msg, s)
		return
	}

	switch msg.Op {
	case pb.MsgServiceInfo_OpDel:
		conns.Delete(makeSesDesc(msg.Name, msg.Id))
		service.Update(msg.Name, msg.Id, msg.EndPoint, msg.Op)
	case pb.MsgServiceInfo_OpUpdate:
		for _, v := range msg.All {
			service.Update(v.Type, v.Id, v.EndPoint, pb.MsgServiceInfo_OpUpdate)
		}
		//service.Update(msg.Name, msg.Id, msg.EndPoint, msg.Op)
	case pb.MsgServiceInfo_OpReqAll:
		for _, v := range msg.All {
			service.Update(v.Type, v.Id, v.EndPoint, pb.MsgServiceInfo_OpUpdate)
		}
	case pb.MsgServiceInfo_OpDelRet:
		if msg.Name == localName && msg.Id == localId {
			<-waitClose
		}
	}
}

func onHeartbeat(msg proto.Message, s *Session) {
	ses := GetSession(s.svcName, s.id) //避免外层逻辑找不到session
	if ses != nil {
		ses.Send(pb.MsgIDS2S_SvcHeartbeatRet, msg)
	}
}

func onHeartbeatRet(msg proto.Message, s *Session) {
	conn, ok := conns.Load(s.desc)
	if ok && conn.(*Conn) != nil {
		atomic.StoreInt64(&conn.(*Conn).heartbeat, time.Now().Unix())
	}
}
