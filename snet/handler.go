package snet

import (
	"google.golang.org/protobuf/proto"
	"server/com/log"
	"server/com/util"
	"server/pb"
	"sync/atomic"
)

var handler = newRoute(int(pb.MsgIDS2S_S2SMsgIDMax))
var netStart uint32

// RegisterCliHandle 注册客户端发来的消息处理函数
func RegisterHandle(msgID pb.MsgIDS2S, cf func() proto.Message, df func(msg proto.Message, s *Session)) {
	if atomic.LoadUint32(&netStart) == 1 {
		log.Errorf("register msg handle %d[%s] failed, you mast register befor StartServe or Connect\n, stack:%s",
			msgID, pb.MsgIDS2S_name[int32(msgID)], util.FuncCaller(3))
		return
	}
	handler.register(uint16(msgID), cf, df)
}

type msgHandler struct {
	createFunc func() proto.Message
	handleFunc func(msg proto.Message, s *Session)
}

// route 消息处理器
type route struct {
	handlers []*msgHandler
}

// newRoute createRoute
func newRoute(size int) *route {
	r := &route{
		make([]*msgHandler, size),
	}
	return r
}

// register 注册消息
func (r *route) register(msgID uint16, cf func() proto.Message, df func(msg proto.Message, s *Session)) {
	n := &msgHandler{
		createFunc: cf,
		handleFunc: df,
	}
	r.handlers[msgID] = n
}

// handle 处理消息
func (r *route) handle(id uint16, data []byte, s *Session) error {
	node, err := r.getHandler(id)
	if err != nil {
		log.Warnf("%s can not find msg %d handle", s.desc, id)
		return nil //服务器间不断开
	}

	msg, err := r.parseMsg(node, data)
	if err != nil {
		log.Warnf("%s parse msg %d %s err:%v", s.desc, id, pb.MsgIDS2S_name[int32(id)], err)
		return err
	}

	if IsTraceProto() && id > uint16(TraceMsgBegin) {
		log.Infof("recv from %s [%d]%s %v", s.desc, id, pb.MsgIDS2S_name[int32(id)], msg)
	}

	node.handleFunc(msg, s)

	return nil
}

func (r *route) getHandler(id uint16) (n *msgHandler, err error) {
	if int(id) >= len(r.handlers) {
		err = errAPINotFind
		return
	}

	n = r.handlers[id]
	if nil == n || nil == n.createFunc {
		err = errAPINotFind
		return
	}
	return
}

func (r *route) parseMsg(n *msgHandler, data []byte) (msg proto.Message, err error) {
	msg = n.createFunc()
	if msg == nil { //允许只有消息id没内容
		return
	}
	err = proto.Unmarshal(data, msg)
	return
}
