package snet

import (
	"context"
	"google.golang.org/protobuf/proto"
	"server/com/log"
	"server/com/util"
	"server/pb"
	"sync"
)

type IStream interface {
	Context() context.Context
	Send(*pb.SrvMsgs) error
	Recv() (*pb.SrvMsgs, error)
}

const (
	SendChanSize  = 4096
	RecvChanSize  = 4096
	MsgCacheSize  = 4096
	TraceMsgBegin = pb.MsgIDS2S_S2STraceEnd
)

// Session 网络会话
type Session struct {
	stream    IStream
	recvChan  chan *pb.SrvMsgs
	sendChan  chan *pb.SrvMsg
	desc      string
	svcName   string
	sn        string
	id        uint32
	die       chan struct{}
	close     chan struct{}
	wait      sync.WaitGroup
	closeSync chan bool
}

// NewSession	新建一个网络会话
func NewSession(svcName string, id uint32, stream IStream) *Session {
	s := &Session{
		stream:    stream,
		die:       make(chan struct{}),
		close:     make(chan struct{}),
		svcName:   svcName,
		id:        id,
		desc:      makeSesDesc(svcName, id),
		recvChan:  make(chan *pb.SrvMsgs, RecvChanSize),
		sendChan:  make(chan *pb.SrvMsg, SendChanSize),
		sn:        util.NewUUID(),
		closeSync: make(chan bool, 1),
	}
	s.wait.Add(1)

	return s
}

func makeSesDesc(svcName string, id uint32) string {
	return svcName + "-" + util.Uint32ToString(id)
}

func (s *Session) ID() uint32 {
	return s.id
}

func (s *Session) Type() string {
	return s.svcName
}

// Send send msg to cli
func (s *Session) Send(msgID pb.MsgIDS2S, msg proto.Message) {
	var b []byte
	var err error

	if msg != nil {
		b, err = proto.Marshal(msg)
		if err != nil {
			log.Warnf("%s marshal err %v when send:%s", s.desc, err, pb.MsgIDS2S_name[int32(msgID)])
			return
		}
	}
	s.SendByte(uint32(msgID), b)

	if IsTraceProto() && msgID > TraceMsgBegin {
		log.Infof("send to %s [%d]%s %s", s.desc, msgID, pb.MsgIDS2S_name[int32(msgID)], msg)
	}
}

func (s *Session) SendByte(msgID uint32, data []byte) {
	s.sendChan <- &pb.SrvMsg{
		Msg: data,
		ID:  msgID,
	}
}

func (s *Session) OnConnect() {
	AddSession(s.svcName, s.id, s)
	s.wait.Done() //必须在GetCallback 前面
	cb := GetCallback(s.svcName, s.id)
	if cb != nil {
		cb(s.id) //不能调用阻塞
		log.Infof("%s %d  %s callback", s.svcName, s.id, s.sn)
	}
	log.Infof("%s[%s] connected success", s.desc, s.sn)
}

// OnClose	离线处理
func (s *Session) OnClose() {
	DelSession(s.svcName, s.id, s.sn)
	s.closeSync <- true
	log.Infof("%s %s closed", s.desc, s.sn)
}

func (s *Session) CloseSync() {
	log.Infof("%s %s sync close", s.desc, s.sn)
	close(s.close)
	<-s.closeSync
	log.Infof("%s %s sync close success", s.desc, s.sn)
}
