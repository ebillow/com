package main

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"server/com/util"
	"strconv"
	"strings"
)

var targetCfgPath string
var srcCfgPath string

func main() {
	targetCfgPath = "./cfgs"
	srcCfgPath = "./src"
	walkFiles(srcCfgPath)
}

func walkFiles(path string) {
	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
		return
	}
	msgs := make([]string, 0)
	for _, d := range dir {
		if d.IsDir() {
			walkFiles(path + "/" + d.Name())
		} else {
			//info, err := d.Info()
			//if err != nil {
			//	log.Fatal(err)
			//}
			//fmt.Println(info)
			msg, err := doOneFile(path, d.Name())
			if err != nil {
				log.Fatal(err)
			}
			msgs = append(msgs, msg...)
		}
	}
	err = writeProtosFile("./proto/configs.proto", msgs)
	if err != nil {
		log.Fatal("make proto file err:", err)
	}
}

func doOneFile(filePath, fileName string) ([]string, error) {
	msgs := make([]string, 0)
	f, err := xlsx.OpenFile(filePath + "/" + fileName)
	if err != nil {
		return nil, err
	}
	sss, err := f.ToSlice()
	if err != nil {
		return nil, err
	}
	for i, sheet := range f.Sheets {
		msgOneSheet, err := doSheet(filePath, sheet.Name, sss[i])
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msgOneSheet)
	}
	return msgs, nil
}

const (
	DescRow      = 0
	FiledNameRow = 1
	FiledTypeRow = 2
)

func doSheet(filePath, fileName string, ss [][]string) (string, error) {
	if len(ss) < 3 {
		return "", fmt.Errorf("%s/%s 表数据空", filePath, fileName)
	}

	head := ss[FiledNameRow]
	types := ss[FiledTypeRow]

	path := strings.TrimLeft(filePath, srcCfgPath)
	f, err := os.OpenFile(targetCfgPath+"/"+path+"/"+fileName+".json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	desc, err := doDesc(ss[DescRow][0])
	if err != nil {
		return "", fmt.Errorf("%s/%s 第一行描述错误", filePath, fileName)
	}

	keys := make(map[string]int, len(ss))

	f.WriteString("{\n")
	//这样做的原因：和excel对应，每一条数据为一行
	for i := FiledTypeRow + 1; i < len(ss); i++ {
		ret, err := doRow(head, types, ss[i])
		if err != nil {
			return "", fmt.Errorf("%s %v", fileName, err)
		}

		b, err := json.Marshal(ret)
		if err != nil {
			return "", fmt.Errorf("%s %v", fileName, err)
		}
		if i > FiledTypeRow+1 {
			f.WriteString(",\n")
		}
		key := makeKey(desc, ret)
		if idx, ok := keys[key]; ok {
			return "", fmt.Errorf("%s[%d and %d]主键[%s]重复", fileName, idx, i, key)
		} else {
			keys[key] = i
		}
		f.WriteString(fmt.Sprintf(`"%s":`, key))
		f.Write(b)
	}
	f.WriteString("\n}")

	msg, err := makeProtos(fileName, head, types)
	if err != nil {
		return "", fmt.Errorf("%s %v", fileName, err)
	}
	return msg, nil
}

func makeKey(keyDesc *Desc, data map[string]interface{}) string {
	key := ""
	for _, v := range keyDesc.Key {
		key += util.ToString(data[v])
	}
	return key
}

func doRow(head, types, s []string) (ret map[string]interface{}, err error) {
	if len(s) != len(head) {
		return nil, fmt.Errorf("%v 行长度和表头不一致", s)
	}
	data := make(map[string]interface{})
	for i := range s {
		v, err := doCell(types[i], s[i])
		if err != nil {
			return nil, fmt.Errorf("%v 行 (%q, %q)解析错误：%v ", s, types[i], s[i], err)
		}
		data[head[i]] = v
	}
	return data, nil
}

func doCell(t string, v string) (ret interface{}, err error) {
	switch t {
	case "string":
		ret = v
	case "int32":
		ret, err = strconv.ParseInt(v, 10, 32)
	case "uint32":
		ret, err = strconv.ParseUint(v, 10, 32)
	case "int64":
		ret, err = strconv.ParseInt(v, 10, 64)
	case "uint64":
		ret, err = strconv.ParseUint(v, 10, 64)
	case "double":
		ret, err = strconv.ParseFloat(v, 64)
	case "bool":
		ret, err = strconv.ParseBool(v)
	default:
		ret = make(map[string]interface{})
		err = json.Unmarshal([]byte(v), &ret)
	}
	return ret, err
}

type Desc struct {
	Key  []string `json:"key"`
	Name string
}

func doDesc(desc string) (*Desc, error) {
	d := &Desc{}
	if len(desc) == 0 {
		return d, nil
	}
	err := json.Unmarshal([]byte(desc), d)
	return d, err
}
