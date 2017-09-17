package main

import (
	"github.com/lvfeiyang/proxy/common/flog"
	"github.com/lvfeiyang/proxy/message"
	"github.com/lvfeiyang/proxy/common/config"
	"net/http"
	"net/rpc"
	"net"
	"fmt"
	"regexp"
	"strings"
)

func main() {
	flog.Init()
	config.Init()

	//http
	http.HandleFunc("/", uiProxyHandler)
	// http.HandleFunc("/msg/", msgProxyHandler)
	http.Handle("/msg/", &message.Message{})
	flog.LogFile.Fatal(http.ListenAndServe(":80", nil))

	//tcp
	ln, err := net.Listen("tcp", ":7777")
	if err != nil {
		flog.LogFile.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			flog.LogFile.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}
func uiProxyHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile("/([^/]+)/?")
	if ur := re.FindStringSubmatch(r.URL.Path); ur != nil {
		pjcCfg := config.GetProjectConfig(ur[1])
		if "" != pjcCfg.Name {
			if rpcCli, err := rpc.Dial("tcp", pjcCfg.Tcp); err != nil {
				flog.LogFile.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				mux := http.NewServeMux()
				if err := rpcCli.Call(strings.Title(pjcCfg.Name)+".Ui", byte(1), mux); err != nil {
					flog.LogFile.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					mux.ServeHTTP(w, r)
				}
			}
		} else {
			fmt.Fprintf(w, "unknow url !!!")
		}
	} else {
		fmt.Fprintf(w, "error url !!!")
	}
}
func handleConnection(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)
	if len, err := conn.Read(buff); err != nil {
		flog.LogFile.Println(err)
	} else {
		msg := &message.Message{}
		if msg.Decode(buff[:len]) != nil {
			flog.LogFile.Println(err)
		} else {
			sendMsg := msg.HandleMsg()
			if "" != sendMsg.Data {
				if sendData, err := sendMsg.Encode(); err != nil {
					flog.LogFile.Println(err)
				} else {
					conn.Write(sendData)
				}
			}
		}
	}
}
