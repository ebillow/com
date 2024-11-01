package snet

import (
	"io"
	"server/com/log"
	"server/com/util"
	"server/pb"
	"time"
)

func recvLoop(s *Session, stream IStream) {
	defer func() {
		log.Debugf("%s close recv loop", s.desc)
		close(s.recvChan)
	}()
	for {
		msg, err := stream.Recv()
		if err == io.EOF { //cli closed
			log.Debugf("close by remote %s", s.desc)
			return
		}
		if err != nil {
			log.Errorf("%s close %v", s.desc, err)
			return
		}
		select {
		case s.recvChan <- msg:
		case <-s.die:
			return
		case <-ctrl:
			return
		}
	}
}

func sendLoop(s *Session, stream IStream, desc string, die chan struct{}) {
	t := time.NewTicker(time.Millisecond * 50)
	cacheMsgs := &pb.SrvMsgs{Msgs: make([]*pb.SrvMsg, 0, MsgCacheSize)}
	defer func() {
		t.Stop()
		log.Debugf("%s close send loop", desc)
	}()
	for {
		select {
		case out := <-s.sendChan:
			cacheMsgs.Msgs = append(cacheMsgs.Msgs, out)
		case <-t.C:
			if len(cacheMsgs.Msgs) != 0 {
				if err := stream.Send(cacheMsgs); err != nil {
					log.Warnf("%s send msg err:%v", desc, err)
				}
				cacheMsgs.Msgs = make([]*pb.SrvMsg, 0, MsgCacheSize)
			}
		case <-ctrl:
			return
		case <-die:
			return
		}
	}
}

func handLoop(s *Session) error {
	s.OnConnect()
	defer func() {
		if err := recover(); err != nil {
			util.PrintStack(err)
		}
	}()

	defer func() {
		s.OnClose()
		close(s.die)
		log.Debugf("%s close handle loop", s.desc)
	}()
	for {
		select {
		case msgs, ok := <-s.recvChan:
			if !ok {
				return nil
			}
			//msgid/2 data
			for _, msg := range msgs.Msgs {
				if err := handler.handle(uint16(msg.GetID()), msg.Msg, s); err != nil {
					log.Warnf("%s hand msg[%d] err:%v", s.desc, msg.GetID(), err)
					return err
				}
			}
		case <-s.close:
			return nil
		case <-ctrl:
			return nil
		}
	}
}
