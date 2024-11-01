package snet

import (
	"google.golang.org/protobuf/proto"
	"server/com/mq"
	"server/pb"
)

var keyMsg *mq.NsqSession

func InitKeyMsg(nsqdAddr string, nsqLkup []string) {
	keyMsg = mq.NewNsqSession(nsqdAddr, nsqLkup, 0)
}

// SendKeyMsg 发送proto buffer消息
func SendKeyMsg(topic string, msgId pb.MsgIDS2S, msgData proto.Message, serId uint32) error {
	return keyMsg.Send(topic, msgId, msgData, serId)
}

// SendKeyMsgData	发送二进制数据
func SendKeyMsgData(topic string, msgId pb.MsgIDS2S, b []byte, serId uint32) error {
	return keyMsg.SendData(topic, msgId, b, serId)
}

// RegisterKeyMsgHandle	注册消息处理函数
func RegisterKeyMsgHandle(msgID pb.MsgIDS2S, cf func() proto.Message, df func(msg proto.Message, serId uint32)) {
	keyMsg.RegisterMsgHandle(msgID, cf, df)
}

// AddConsumer	添加一个消费者
func AddConsumer(topic string, chl string) {
	_ = SendKeyMsg(topic, pb.MsgIDS2S_MsgIDS2SNone, nil, 0)
	keyMsg.AddConsumer(topic, chl)
}
