package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/com/config/pb"
	"server/com/log"
	"server/com/util"
	"strings"
	"unicode/utf8"
)

var cfgAll = &pb.Cfg{}
var fileData strings.Builder
var first = true

func ReadAll() *pb.Cfg {
	fileData.WriteString("{")
	readDir("./cfgs", "")
	fileData.WriteString("}")
	str := fileData.String()
	fmt.Sprintln(str)
	err := json.Unmarshal([]byte(str), cfgAll)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cfgAll)
	return cfgAll
}

func readDir(path string, ex string) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Errorf("Read Configs error")
	}
	for _, file := range files {
		npath := filepath.Join(path, file.Name())
		if file.IsDir() {
			nex := ex + util.Title(file.Name())
			readDir(npath, nex)
		} else {
			ReadOneFile(file, npath, ex)
		}
	}
}

func ReadOneFile(file os.DirEntry, npath string, ex string) error {
	ext := filepath.Ext(file.Name())
	fi := strings.TrimSuffix(file.Name(), ext)
	nex := ex + util.Title(fi)
	data, err := readData(npath)
	if err != nil {
		return err
	}
	if !first {
		fileData.WriteString(",")
	} else {
		first = false
	}
	fileData.WriteString(fmt.Sprintf(`"%s":`, nex))
	fileData.Write(data)
	return nil
}

func readData(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	if !utf8.Valid(byteValue) {
		return nil, fmt.Errorf("file is not utf8")
	}

	return byteValue, nil
}
