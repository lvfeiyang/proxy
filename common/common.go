package common

import (
	"github.com/lvfeiyang/proxy/common/flog"
	"github.com/lvfeiyang/proxy/message"
	"net"
	"regexp"
)

func ListenTcp(addr string, mmh message.MsgMapHandle) {
	ln, err := net.Listen("tcp", addr)
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
		go handleConnection(conn, mmh)
	}
}
func handleConnection(conn net.Conn, mmh message.MsgMapHandle) {
	defer conn.Close()
	buff := make([]byte, 1024)
	if len, err := conn.Read(buff); err != nil {
		flog.LogFile.Println(err)
	} else {
		msg := &message.Message{}
		if msg.Decode(buff[:len]) != nil {
			flog.LogFile.Println(err)
		} else {
			sendMsg := msg.HandleMsg(mmh)
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

func ImgUrlAddQn(img string) string {
	domainMapUrl := map[string]string{
		"xiaobai": "http://ov4dqx58l.bkt.clouddn.com",
		"photo": "http://oz6l8jt3v.bkt.clouddn.com",
	}
	re := regexp.MustCompile("(.*?)/")
	imgreg := re.FindStringSubmatch(img)
	if imgreg != nil {
		if url, ok := domainMapUrl[imgreg[1]]; ok {
			return string(re.ReplaceAll([]byte(img), []byte(url+"/"))) // + "?imageView2/4/w/300/h/300"
		}
	}
	return img
}
