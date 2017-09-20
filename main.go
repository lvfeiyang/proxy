package main

import (
	"fmt"
	"github.com/lvfeiyang/proxy/common/config"
	"github.com/lvfeiyang/proxy/common/flog"
	"github.com/lvfeiyang/proxy/message"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"regexp"
)

func main() {
	flog.Init()
	config.Init()
	htmlPath := config.ConfigVal.HtmlPath

	//http
	jsFiles := filepath.Join(htmlPath, "sfk", "js")
	cssFiles := filepath.Join(htmlPath, "sfk", "css")
	fontsFiles := filepath.Join(htmlPath, "sfk", "fonts")
	layDateFiles := filepath.Join(htmlPath, "sfk", "laydate")
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(jsFiles))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(cssFiles))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir(fontsFiles))))
	http.Handle("/laydate/", http.StripPrefix("/laydate/", http.FileServer(http.Dir(layDateFiles))))

	//tcp
	go proxyTcp()

	http.HandleFunc("/", proxyHandler)
	flog.LogFile.Fatal(http.ListenAndServe(":80", nil))
}
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile("/([^/]+)/?")
	if ur := re.FindStringSubmatch(r.URL.Path); ur != nil {
		if pjtCfg := config.GetProjectConfig(ur[1]); "" != pjtCfg.Name {
			re := regexp.MustCompile("/([^/]+)/msg/(.+)")
			if ur := re.FindStringSubmatch(r.URL.Path); ur != nil {
				//message
				msg := &message.Message{}
				msg.FromHttp(r)
				sendMsg := msg.SendToInside(pjtCfg.Tcp)
				sendMsg.ToHttp(w)
			} else {
				//ui
				rpu, err := url.Parse("http://" + pjtCfg.Http)
				if err != nil {
					flog.LogFile.Fatal(pjtCfg.Name, "http addr error ", err)
				}
				rp := httputil.NewSingleHostReverseProxy(rpu)
				rp.ServeHTTP(w, r)
			}
		} else {
			fmt.Fprintf(w, "unknow url !!!")
		}
	} else {
		fmt.Fprintf(w, "error url !!!")
	}
}
func proxyTcp() {
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
		go proxyHandleConn(conn)
	}
}
func proxyHandleConn(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)
	if len, err := conn.Read(buff); err != nil {
		flog.LogFile.Println(err)
	} else {
		msg := &message.Message{}
		if msg.Decode(buff[:len]) != nil {
			flog.LogFile.Println(err)
		} else {
			sendMsg := msg.SendToInside("")
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
