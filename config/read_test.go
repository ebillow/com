package main

import (
	"encoding/json"
	"fmt"
	"os"
	"server/com/config/pb"
	"strconv"
	"testing"
)

func TestReadOneFile(t *testing.T) {
	b, err := os.ReadFile("./cfgs/Prize.json")
	if err != nil {
		return
	}
	cfg := &pb.Cfg{}
	err = json.Unmarshal(b, &cfg.Prize)
	if err != nil {
		return
	}
	t.Log(cfg)
}

func TestWriteOneFile(t *testing.T) {
	cfg := &pb.Cfg{}
	cfg.Prize = make(map[string]*pb.CfgPrize)
	for i := 0; i < 3; i++ {
		cfg.Prize[strconv.Itoa(i)] = &pb.CfgPrize{
			Id:    int32(i),
			Prize: []int32{1, 2},
			Item:  &pb.ItemCnt{Cnt: 1, Id: 2},
		}
	}
	b, err := json.Marshal(cfg.Prize)
	if err != nil {
		fmt.Println(err)
		return
	}
	os.WriteFile("./cfgs/Prize.json", b, 0644)
}

func TestWriteAll(t *testing.T) {
	cfg := &pb.Cfg{}
	cfg.Prize = make(map[string]*pb.CfgPrize)
	for i := 1; i < 4; i++ {
		cfg.Prize[strconv.Itoa(i)] = &pb.CfgPrize{
			Id:    int32(i),
			Prize: []int32{1, 2, 2 + int32(i)},
			Item:  &pb.ItemCnt{Cnt: int64(i) + 1, Id: 1},
		}
	}

	cfg.Test = make(map[string]*pb.CfgTest)
	for i := int32(1); i < 4; i++ {
		cfg.Test[strconv.Itoa(int(i))] = &pb.CfgTest{
			Id:    int32(i),
			Prize: []int32{1, 2, 2 + i},
			Item:  &pb.ItemCnt{Cnt: int64(i) + 1, Id: 1},
			Gift:  1,
			Name:  "1",
		}
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}

func TestReadAll(t *testing.T) {
	ReadAll()
}
