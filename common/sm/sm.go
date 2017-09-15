package sm

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	// "net/http/httputil"
	"github.com/lvfeiyang/guild/common/flog"
	"net/url"
	"strings"
	"time"
	// "crypto/tls"
	// "strconv"
	"fmt"
)

const (
	accountSid = "8a216da85bf14b6a015bff36d51d058b"
	accountId  = "8a216da85bf14b6a015bff36d51d058b"
	authToken  = "a24dcbf3a0dd4a648cb06871602b06ce"
)

var tidBelongAid = map[string]string{
	"176875": "8a216da85bf14b6a015bff36d5bc058f",
}

type smReq struct {
	To         string   `json:"to"`
	AppId      string   `json:"appId"`
	TemplateId string   `json:"templateId"`
	Datas      []string `json:"datas"`
}

type rsp1 struct {
	DateCreated   string `json:"dateCreated"`
	SmsMessageSid string `json:"smsMessageSid"`
}
type smRsp struct {
	StatusCode  string `json:"statusCode"`
	TemplateSMS rsp1   `json:"templateSMS"`
}

func sendSM(mobile, templateId string, datas []string) {
	//组装url
	now := time.Now()
	u, _ := url.Parse("https://app.cloopen.com:8883")
	u.Opaque = "/2013-12-26/Accounts/" + accountSid + "/SMS/TemplateSMS"
	q := u.Query()
	signed := accountId + authToken + now.Format("20060102150405")
	h := md5.Sum([]byte(signed))
	sign := fmt.Sprintf("%X", h)
	q.Set("sig", sign)
	u.RawQuery = q.Encode()

	//组装请求body
	reqBody := smReq{mobile, tidBelongAid[templateId], templateId, datas}
	reqBodyj, err := json.Marshal(reqBody)
	if err != nil {
		flog.LogFile.Println(err)
	}
	body := bytes.NewReader(reqBodyj)

	req, err := http.NewRequest("POST", "", body)
	if err != nil {
		flog.LogFile.Println(err)
	}
	req.URL = u

	//组装请求header
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	authen := base64.URLEncoding.EncodeToString([]byte(accountId + ":" + now.Format("20060102150405")))
	req.Header.Set("Authorization", authen)

	// dump, _ := httputil.DumpRequest(req, true)
	// flog.LogFile.Println(string(dump))

	// tr := &http.Transport{TLSClientConfig:&tls.Config{InsecureSkipVerify:true}}
	// client := &http.Client{Transport: tr}
	rsp, err := http.DefaultClient.Do(req) //http.DefaultClient
	if err != nil {
		flog.LogFile.Println(err)
	}

	//解析响应
	defer rsp.Body.Close()
	rspBodyj, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		flog.LogFile.Println(err)
	}
	rspBody := &smRsp{}
	if err := json.Unmarshal(rspBodyj, rspBody); err != nil {
		flog.LogFile.Println(err)
	}
	if 0 != strings.Compare("000000", rspBody.StatusCode) {
		flog.LogFile.Println("yuntongxun send msg fail! error code: ", rspBody.StatusCode)
	}
}

func SendVerifyCode(mobile, code string) {
	go sendSM(mobile, "176875", []string{code})
}
