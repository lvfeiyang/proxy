package flog

import (
	"log"
	"os"
	"runtime"
)

var LogFile = &log.Logger{}

func Init() {
	var filePath string
	if "linux" == runtime.GOOS {
		filePath = "/root/workspace/xiaobai/log"
	} else {
		filePath = "C:\\Users\\Administrator\\log"
	}
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	LogFile = log.New(file, "", log.LstdFlags) //error:
}
