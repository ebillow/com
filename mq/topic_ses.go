package mq

import (
	"google.golang.org/protobuf/proto"
	"server/com/log"
	"server/pb"
	"time"
)

const SendSize = 4096

type TopicAsyncSession struct {
	*NsqSession
	topic string
	c     chan []byte
	cache [][]byte
}

// NewTopicAsyncSession 指定主题会话，消息每隔hzMSec毫秒异步发送
func NewTopicAsyncSession(logTopic string, nsqdAddr string, nsqLkup []string, hzMSec int64) *TopicAsyncSession {
	m := &TopicAsyncSession{
		NsqSession: NewNsqSession(nsqdAddr, nsqLkup, 0),
		c:          make(chan []byte, SendSize),
		cache:      make([][]byte, 0, SendSize),
		topic:      logTopic,
	}
	if hzMSec == 0 {
		hzMSec = 1000
	}
	if hzMSec < 50 {
		log.Infof("hzMSec max >= 50,change it to 50")
		hzMSec = 50
	}
	go m.run(hzMSec)
	return m
}

func (n *TopicAsyncSession) Send(msgId pb.MsgIDS2S, msgData proto.Message, serId uint32) error {
	data, err := writePkg(uint16(msgId), msgData, serId)
	if err != nil {
		return err
	}
	n.c <- data

	return nil
}

func (n *TopicAsyncSession) run(hzMSec int64) {
	t := time.NewTicker(time.Millisecond * time.Duration(hzMSec))
	//doneChan := make(chan *nsq.ProducerTransaction, SendSize)
	var err error
	for {
		select {
		case b := <-n.c:
			n.cache = append(n.cache, b)
		case <-t.C:
			if len(n.cache) > 0 {
				err = n.producer.MultiPublishAsync(n.topic, n.cache, nil)
				if err != nil {
					log.Warnf("send log err:%v", err)
				}
				n.cache = make([][]byte, 0, SendSize)
			}
		}
	}
}
