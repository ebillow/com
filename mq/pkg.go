package mq

import (
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"server/com/log"
)

func writePkg(msgId uint16, msg proto.Message, serId uint32) ([]byte, error) {
	var b []byte
	var err error

	if msg != nil {
		b, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	}

	return writeData(msgId, b, serId)
}

func writeData(msgId uint16, b []byte, serId uint32) ([]byte, error) {
	msgLen := len(b) + 6
	data := make([]byte, msgLen)
	binary.BigEndian.PutUint16(data[0:2], msgId)
	binary.BigEndian.PutUint32(data[2:6], serId)
	copy(data[6:], b)
	return data, nil
}

func parserPkg(data []byte) (msgId uint16, msg []byte, serId uint32) {
	if len(data) < 6 {
		log.Errorf("server recv data len < 6 %v", data)
		return 0, nil, 0
	}
	msgId = binary.BigEndian.Uint16(data[0:2])
	serId = binary.BigEndian.Uint32(data[2:6])
	msg = data[6:]
	return
}
