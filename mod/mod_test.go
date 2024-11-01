package mod

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"server/pb"
	"testing"
	"time"
)

type ModTest struct {
}

func (mt *ModTest) OnProto(message proto.Message, ctxWithValue context.Context) {
	fmt.Println(message, ctxWithValue.Value(1))
}

func (mt *ModTest) Save() {

}

func TestMod(b *testing.T) {
	err := Register(1, &ModTest{})
	if err != nil {
		b.Fatal()
	}
	Start()
	for i := 0; i < 5; i++ {
		Post(1, &pb.AccInfo{Acc: "acc"}, context.WithValue(context.Background(), 1, i))
	}
	select {
	case <-time.After(5 * time.Second):
	}
	Stop()
}
