package message

import (
	"encoding/hex"
	"encoding/json"
	"github.com/lvfeiyang/proxy/common/config"
	"github.com/lvfeiyang/proxy/common/flog"
	"github.com/lvfeiyang/proxy/common/session"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Message struct {
	Name      string
	Data      string
	SessionId uint64
	Project   string
}

func (msg *Message) FromHttp(r *http.Request) {
	re := regexp.MustCompile("/([^/]+)/msg/(.+)")
	if ur := re.FindStringSubmatch(r.URL.Path); ur != nil {
		msg.Project = ur[1]
		msg.Name = ur[2] + "-req"
	}

	// msg.Name = r.URL.Path[len("/msg/"+project+"/"):] + "-req"
	var err error
	if headSessId := r.Header.Get("SessionId"); "" == headSessId {
		msg.SessionId = 0
	} else {
		msg.SessionId, err = strconv.ParseUint(headSessId, 10, 64)
		if err != nil {
			flog.LogFile.Println(err)
		}
	}
	if 0 == strings.Compare("application/json", r.Header.Get("Content-Type")) {
		defer r.Body.Close()
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			flog.LogFile.Println(err)
		}
		msg.Data = string(buff)
	} else {
		// IDEA: form表单需整合为json
		flog.LogFile.Println("http body not json")
	}
}
func (msg *Message) ToHttp(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if 0 == strings.Compare("error-msg", msg.Name) {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(msg.Data))
}

func (msg *Message) SendToInside(tcpAddr string) *Message {
	insMsg := &Message{Name: "error-msg", Data: UnknowError()}
	if "" == tcpAddr {
		if pjtCfg := config.GetProjectConfig(msg.Project); "" != pjtCfg.Name {
			tcpAddr = pjtCfg.Tcp
		} else {
			// flog.LogFile.Println("unknow project")
			insMsg = &Message{Name: "error-msg", Data: UnknowMsg()}
		}
	}

	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		flog.LogFile.Println("cann't connect tcp", tcpAddr)
	}
	defer conn.Close()

	if send, err := msg.Encode(); err != nil {
		flog.LogFile.Println(err)
	} else {
		if _, err := conn.Write(send); err != nil {
			flog.LogFile.Println(err)
		}
		recvData := make([]byte, 8*1024)
		if n, err := conn.Read(recvData); err != nil {
			flog.LogFile.Println(err)
		} else {
			insMsg.Decode(recvData[:n])
		}
	}
	return insMsg
}

func GeneralServeHTTP(msg *Message, w http.ResponseWriter, r *http.Request, mmh MsgMapHandle) {
	msg.FromHttp(r)
	sendMsg := msg.HandleMsg(mmh)
	sendMsg.ToHttp(w)
}

func (msg *Message) Decode(data []byte) error {
	return json.Unmarshal(data, msg)
}
func (msg *Message) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

type MsgHandleIF interface {
	Decode(msgData []byte) error
	Handle(sess *session.Session) ([]byte, error)
	GetName() (string, string)
}
type MsgMapHandle map[string]MsgHandleIF

func deCrypto(msgData []byte, sess *session.Session) ([]byte, error) {
	recvEn := make([]byte, hex.DecodedLen(len(msgData)))
	n, err := hex.Decode(recvEn, msgData)
	if err != nil {
		return nil, err
	}
	recv, err := AesDe(recvEn[:n], NewKey(sess.N))
	if err != nil {
		return nil, err
	}
	return recv, nil
}
func handleOneMsg(req MsgHandleIF, msgData []byte, sess *session.Session) *Message {
	sendMsg := &Message{Name: "error-msg", Data: UnknowError()}
	reqName, rspName := req.GetName()

	if req.Decode(msgData) != nil {
		sendMsg = &Message{Name: "error-msg", Data: DecodeError(reqName)}
	} else {
		var rspData []byte
		var err interface{}
		// req.SessionId = msgSessId
		rspData, err = req.Handle(sess)
		if err != nil {
			if _, ok := err.(*ErrorMsg); ok {
				sendMsg = &Message{Name: "error-msg", Data: string(rspData)}
			} else {
				flog.LogFile.Println(err)
			}
		} else {
			sendMsg = &Message{Name: rspName, Data: string(rspData)}
		}
	}
	return sendMsg
}

func (msg *Message) HandleMsg(mmh MsgMapHandle) *Message {
	sess := &session.Session{SessId: msg.SessionId}
	if 0 != msg.SessionId {
		if err := sess.Get(msg.SessionId); err != nil {
			errData, _ := NormalError(ErrGetSessionFail)
			return &Message{Name: "error-msg", Data: string(errData)}
		}
	}
	msgIF, ok := mmh[msg.Name]
	if !ok {
		return &Message{Name: "error-msg", Data: UnknowMsg()}
	}
	var msgData []byte
	if needCrypto(msg.Name, msg.Project) {
		var err error
		msgData, err = deCrypto([]byte(msg.Data), sess)
		if err != nil {
			errData, _ := NormalError(ErrDeCrypto)
			return &Message{Name: "error-msg", Data: string(errData)}
		}
	} else {
		msgData = []byte(msg.Data)
	}
	return handleOneMsg(msgIF, msgData, sess)
}

func needCrypto(name, project string) bool {
	return false
}
