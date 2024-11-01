package util

import (
	"fmt"
	"runtime"
	"server/com/ding"
	"server/com/log"
)

//PrintStack 打印堆栈，并打印传入的变量
func PrintStack(vars ...interface{}) {
	stack := make([]string, 0)
	for _, v := range vars {
		stack = append(stack, fmt.Sprintf("%v\n", v))
	}

	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	stack = append(stack, string(buf[0:n]))

	log.Error(stack)

	if !Debug {
		var msg string
		for i := 0; i < len(stack); i++ {
			msg += stack[i]
			msg += "\n"
		}
		ding.SendDing(msg)
	}
}

//FuncCaller 得到调用者
func FuncCaller(lvl int) string {
	funcName, file, line, ok := runtime.Caller(lvl)

	info := ""
	for ok {
		info += fmt.Sprintf("frame %v:[func:%v,file:%v,line:%v]\n", lvl, runtime.FuncForPC(funcName).Name(), file, line)
		lvl++
		funcName, file, line, ok = runtime.Caller(lvl)
	}

	return info
}

func FuncCallerOnce(lvl int) string {
	_, file, line, ok := runtime.Caller(lvl)
	if ok {
		return fmt.Sprintf("file:%v,line:%v", file, line)
	}
	return ""
}
