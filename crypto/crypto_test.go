package crypto

import (
	"crypto/rc4"
	"google.golang.org/protobuf/proto"
	"server/pb"
	"testing"
)

var data []byte
var key []byte
var msg *pb.MsgPlayerData

func initData() {
	msg = &pb.MsgPlayerData{

		Level: 5,
		Exp:   6,
	}

	key = []byte("1234561111111111")
	data, _ = proto.Marshal(msg)
}

func Benchmark_Rc4(t *testing.B) {
	initData()

	c1, err := rc4.NewCipher(key)
	if err != nil {
		t.Fatal(err)
		return
	}
	c2, err := rc4.NewCipher(key)
	if err != nil {
		t.Fatal(err)
		return
	}
	for i := 0; i != t.N; i++ {
		c1.XORKeyStream(data, data)
		c2.XORKeyStream(data, data)
	}

	msg2 := &pb.MsgPlayerData{}
	err = proto.Unmarshal(data, msg2)
	if err != nil {
		t.Fatal("unmarshal err", err)
		return
	}

	if !proto.Equal(msg, msg2) {
		t.Fatal("not same")
		return
	}
}

func Test_Rc4(t *testing.T) {
	initData()

	c1, err := rc4.NewCipher(key)
	if err != nil {
		t.Fatal(err)
		return
	}
	c2, err := rc4.NewCipher(key)
	if err != nil {
		t.Fatal(err)
		return
	}
	for i := 0; i != 1000000; i++ {
		c1.XORKeyStream(data, data)
		c2.XORKeyStream(data, data)
	}

	msg2 := &pb.MsgPlayerData{}
	err = proto.Unmarshal(data, msg2)
	if err != nil {
		t.Fatal("unmarshal err", err)
		return
	}

	if !proto.Equal(msg, msg2) {
		t.Fatal("not same")
		return
	}
	t.Log(msg)
	t.Log(msg2)
}
