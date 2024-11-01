package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func ReadJson(cfg interface{}, fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, cfg)
	if err != nil {
		return err
	}
	return nil
}

func GetComputerIp() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addrStr := make([]string, 0)
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				addrStr = append(addrStr, ipnet.IP.String())
			}
		}
	}
	return addrStr
}

func GetMd5(message string) string {
	md5hash := md5.New()
	md5hash.Write([]byte(message))
	md5byte := md5hash.Sum(nil)
	hashcode := hex.EncodeToString(md5byte[:])
	return hashcode
}
