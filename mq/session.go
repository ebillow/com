package mq

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"server/com/log"
	"server/com/util"
	"server/pb"
	"sync/atomic"
	"time"
)

type consumerInfo struct {
	consumer *nsq.Consumer
	topic    string
	chanel   string
}

// NsqSession	服务器间消息会话
type NsqSession struct {
	handler   *routeNsq
	producer  *nsq.Producer
	cfg       *nsq.Config
	lkupAddr  []string
	consumers []*consumerInfo
}

func NewNsqSession(nsqdAddr string, nsqLkup []string, cnt int) *NsqSession {
	m := &NsqSession{
		handler: newRouteCt(int(pb.MsgIDS2S_S2SMsgIDMax)),
	}

	m.cfg = nsq.NewConfig()
	m.cfg.LookupdPollInterval = time.Second * 5
	if cnt != 0 {
		m.cfg.MaxInFlight = cnt
	}
	//producer
	m.addProducer(nsqdAddr)

	//consumer
	m.lkupAddr = nsqLkup

	return m
}

// Send 发送proto buffer消息
func (n *NsqSession) Send(topic string, msgId pb.MsgIDS2S, msgData proto.Message, serId uint32) error {
	data, err := writePkg(uint16(msgId), msgData, serId)
	if err != nil {
		return err
	}
	err = n.producer.Publish(topic, data)
	if err != nil {
		return err
	}
	if IsTraceProto() && msgId > pb.MsgIDS2S_S2STraceEnd {
		log.Infof("send to %s [%d][%s] %v", topic, msgId, pb.MsgIDS2S_name[int32(msgId)], msgData)
	}

	return nil
}

// SendData	发送二进制数据
func (n *NsqSession) SendData(topic string, msgId pb.MsgIDS2S, b []byte, serId uint32) error {
	data, err := writeData(uint16(msgId), b, serId)
	if err != nil {
		return err
	}
	err = n.producer.Publish(topic, data)
	if err != nil {
		return err
	}

	return nil
}

// Stop	停止
func (n *NsqSession) Stop() {
	//if n.producer != nil {
	//	n.producer.Stop()
	//}
	for _, v := range n.consumers {
		v.consumer.Stop()
	}
}

func (n *NsqSession) DeleteChannel(adminAddr string, topic string, chl string) {
	client := http.Client{}
	///api/topics/:topic/:channel
	url := fmt.Sprintf("http://%s/api/topics/%s/%s", adminAddr, topic, chl)
	req, _ := http.NewRequest("DELETE", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Warnf("delete [%s]channel:%s err:%v", topic, chl, err)
		return
	}
	_ = resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Warnf("delete [%s]channel:%s err:%s", topic, chl, string(body))
	} else {
		log.Infof("delete [%s]channel:%s", topic, chl)
	}
}

// RegisterMsgHandle	注册消息处理函数
func (n *NsqSession) RegisterMsgHandle(msgID pb.MsgIDS2S, cf func() proto.Message, df func(msg proto.Message, serId uint32)) {
	n.handler.register(msgID, cf, df)
}

// HandleMessage 消息处理入口，nsq自己调用
func (n *NsqSession) HandleMessage(msg *nsq.Message) error {
	defer func() {
		if err := recover(); err != nil {
			util.PrintStack(err)
		}
	}()

	msgId, data, serId := parserPkg(msg.Body)
	err := n.handler.handle(msgId, data, serId)
	if err != nil && msgId > uint16(pb.MsgIDS2S_S2STraceEnd) {
		log.Warnf("handle msg %d err:%v", msgId, err)
	}
	return nil
}

// AddConsumer	添加一个消费者
func (n *NsqSession) AddConsumer(topic string, chl string) {
	n.AddConsumerMulti(topic, chl, 1)
}

// AddConsumerMulti	添加一个消费者,多个协程处理消息
func (n *NsqSession) AddConsumerMulti(topic string, chl string, concurrency int) {
	consumer, err := nsq.NewConsumer(topic, chl, n.cfg)
	if err != nil {
		log.Panic(err)
		return
	}
	consumer.AddConcurrentHandlers(n, concurrency)
	consumer.SetLoggerLevel(nsq.LogLevelWarning)

	err = consumer.ConnectToNSQLookupds(n.lkupAddr)
	if err != nil {
		log.Panic(err)
	}

	n.consumers = append(n.consumers, &consumerInfo{
		consumer: consumer,
		topic:    topic,
		chanel:   chl,
	})
	log.Infof("add consumer t:%s, c:%s to %s", topic, chl, n.lkupAddr)
}

// addProducer	添加一个生产者
func (n *NsqSession) addProducer(addr string) bool {
	var err error
	n.producer, err = nsq.NewProducer(addr, n.cfg)
	if err != nil {
		log.Panic(err)
		return false
	}

	if err = n.producer.Ping(); err != nil {
		log.Warnf("ping NSQ err:%v, wait NSQ ready...", err)
		return false
	}

	log.Infof("add producer to %s", addr)
	return true
}

var isTraceProto int32

func IsTraceProto() bool {
	return atomic.LoadInt32(&isTraceProto) == 1
}

func SetTraceProto(v bool) {
	log.Infof("mq set trace msg :%t", v)
	if v {
		atomic.StoreInt32(&isTraceProto, 1)
	} else {
		atomic.StoreInt32(&isTraceProto, 0)
	}
}
