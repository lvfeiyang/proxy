package main

import (
	"github.com/lvfeiyang/proxy/common/flog"
	// "github.com/lvfeiyang/proxy/message"
	"github.com/lvfeiyang/proxy/common/config"
	"net/http"
	// "net/rpc"
	// "net"
	"fmt"
	"regexp"
	// "strings"
	"path/filepath"
	"net/http/httputil"
	"net/url"
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

	http.HandleFunc("/", uiProxyHandler)
	// http.Handle("/", uiProxyHandler)
	// http.HandleFunc("/msg/", msgProxyHandler)
	// http.Handle("/msg/", &message.Message{})
	flog.LogFile.Fatal(http.ListenAndServe(":80", nil))

	//tcp
	/*ln, err := net.Listen("tcp", ":7777")
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
	}*/
}
func uiProxyHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile("/([^/]+)/?")
	if ur := re.FindStringSubmatch(r.URL.Path); ur != nil {
		if pjtCfg := config.GetProjectConfig(ur[1]); "" != pjtCfg.Name {
			rpu, err := url.Parse("http://"+pjtCfg.Http)
			if err != nil {
				flog.LogFile.Fatal(pjtCfg.Name, "http addr error ", err)
			}
			rp := httputil.NewSingleHostReverseProxy(rpu)
			rp.ServeHTTP(w, r)
		} else {
			fmt.Fprintf(w, "unknow url !!!")
		}
	} else {
		fmt.Fprintf(w, "error url !!!")
	}
}
/*func handleConnection(conn net.Conn) {
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
}*/
