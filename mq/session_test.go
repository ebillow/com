package mq

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
	"server/pb"
	"testing"
	"time"
)

var c = make(chan bool)

func TestSession(t *testing.T) {
	s := NewNsqSession("127.0.0.1:4150", []string{"127.0.0.1:4161"}, 1)

	s.RegisterMsgHandle(pb.MsgIDS2S_MsgIDS2SNone, func() proto.Message { return &pb.MsgStr{} }, onRecvMsg)

	s.AddConsumer("test", "test1")
	s.Send("test", pb.MsgIDS2S_MsgIDS2SNone, &pb.MsgStr{Value: "just test"}, 1)
	c <- true
	close(c)
	s.Stop()
}

func onRecvMsg(msg proto.Message, serID uint32) {
	fmt.Print(msg)
	<-c
}

func TestDelChl(t *testing.T) {
	addr := "http://127.0.0.1:4161/channel/delete"
	data := url.Values{}
	data.Add("topic", "gateAll")
	data.Add("channel", "gt3")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(addr, data)
	if err != nil {
		t.Errorf("Http Post1 Error: %v \n error: %v", addr, err)
		return
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(string(ret))
}

func TestDelChl2(t *testing.T) {
	client := http.Client{}
	topicName := "gateAll"
	channelName := "gt1"
	url := fmt.Sprintf("http://%s/channel/delete?topic=%s&channel=%s", "127.0.0.1:4161", topicName, channelName)

	req, _ := http.NewRequest("POST", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if 200 != resp.StatusCode {
		t.Error(resp.StatusCode)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	t.Logf("%s", body)
}

func TestDelChl3(t *testing.T) {
	client := http.Client{}
	///api/topics/:topic/:channel
	url := fmt.Sprintf("http://%s/api/topics/%s/%s", "127.0.0.1:4171", "gateAll", "gt2")
	req, _ := http.NewRequest("DELETE", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	t.Logf("%s", body)
}
