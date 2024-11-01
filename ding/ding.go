package ding

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"server/com/crypto"
	"time"
)

type MsgContext struct {
	Content string `json:"content"`
}

type MsgAt struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

type MsgDing struct {
	Msgtype string      `json:"msgtype"`
	Text    *MsgContext `json:"text"`
	At      *MsgAt      `json:"at"`
}

type MarkDown struct {
	Title string
	Text  string
}

type MsgDingRet struct {
	Errcode uint32 `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

var webHook = `https://oapi.dingtalk.com/robot/send?access_token=1a8348a76efb3b805363045fb76623c91e8cc500b5cd6da335b1ff75e444dd6e`
var preHead string

func SetPreHead(pre string) {
	preHead += pre + "_"
}

func SendDing(info string) {
	content := fmt.Sprintf("server:%s %s 服务器报警:\n%s", preHead, time.Now().String(), info)
	data := &MsgDing{
		Msgtype: "text",
		Text:    &MsgContext{Content: content},
		At: &MsgAt{
			IsAtAll: true, //@所有人
		},
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("ding marshal err:%v", err)
		return
	}
	ret, err := httpPost(webHook, b, true, true)
	if err != nil {
		log.Printf("ding post err:%v", err)
		return
	}
	retmsg := &MsgDingRet{}
	err = json.Unmarshal(ret, retmsg)
	if err != nil {
		log.Printf("ding unmarshal ret err:%v", err)
	} else {
		log.Print(retmsg)
	}
}

func SendInfo(info string) {
	content := preHead + "服务器信息:" + info
	data := &MsgDing{
		Msgtype: "text",
		Text:    &MsgContext{Content: content},
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("ding marshal err:%v", err)
		return
	}
	ret, err := httpPost(webHook, b, true, true)
	if err != nil {
		log.Printf("ding post err:%v", err)
		return
	}
	retmsg := &MsgDingRet{}
	err = json.Unmarshal(ret, retmsg)
	if err != nil {
		log.Printf("ding unmarshal ret err:%v", err)
	} else {
		log.Print(retmsg)
	}
}

func httpPost(url string, params []byte, isJson bool, isHttps bool) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(params))
	if err != nil {
		log.Printf("Http Post Error: %v \n error: %v", url, err)
		return nil, err
	}
	defer req.Body.Close()

	if isJson {
		req.Header.Add("content-type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	if isHttps {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: crypto.LoadCA()},
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Http Post1 Error: %v \n error: %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
