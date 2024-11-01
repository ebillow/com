package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"text/template"
)

type PbStruct struct {
	StructName string
	Filed      []*PbFiled
}

type PbFiled struct {
	FiledName string
	FiledNO   int
	FiledType string
}

var All = PbStruct{StructName: "Cfg"}
var curNo = 1

func writeProtosFile(filePathName string, msgs []string) error {
	f, err := os.OpenFile(filePathName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			fmt.Println("close file err:", err)
		}
	}()

	_, err = f.Write([]byte(`
syntax = "proto3";
option go_package = "./;pb";
import "common.proto";
	`))

	str, err := structToStr(&All)
	if err != nil {
		return err
	}
	msgs = append(msgs, str)
	for i := range msgs {
		_, err = f.WriteString(msgs[i])
		if err != nil {
			return err
		}
	}

	return err
}

func makeProtos(name string, filedName []string, filedTypes []string) (ret string, err error) {
	if len(filedName) != len(filedTypes) {
		return "", errors.New("len(filedName) != len(filedTypes)")
	}
	data := &PbStruct{
		StructName: "Cfg" + name,
		Filed:      make([]*PbFiled, 0, len(filedName)),
	}
	for i := 0; i < len(filedName); i++ {
		data.Filed = append(data.Filed, &PbFiled{
			FiledName: filedName[i],
			FiledType: filedTypes[i],
			FiledNO:   i + 1,
		})
	}

	All.Filed = append(All.Filed, &PbFiled{
		FiledName: name,
		FiledNO:   curNo,
		FiledType: "map<string, Cfg" + name + ">",
	})
	curNo++

	return structToStr(data)
}

func structToStr(data *PbStruct) (string, error) {
	var msgTemp = `
message {{.StructName}} {
{{range $i, $v := .Filed}}	{{$v.FiledType}}	{{$v.FiledName}}	= {{$v.FiledNO}};
{{end}}}`

	tmpl := template.New("test")
	tmpl = template.Must(tmpl.Parse(msgTemp))
	buff := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buff, data); err != nil {
		return "", err
	}

	return buff.String(), nil
}
