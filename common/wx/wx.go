package wx

import (
	"encoding/json"
	"github.com/lvfeiyang/proxy/common/config"
	"github.com/lvfeiyang/proxy/common/flog"
	"github.com/lvfeiyang/proxy/common/session"
	"io/ioutil"
	"net/http"
)

type accessTokenRsp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint   `json:"expires_in"`
}
type errorRsp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}
type jsapiTicketRsp struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	Ticket    string `json:"ticket"`
	ExpiresIn uint   `json:"expires_in"`
}

func AccessToken() string {
	// var token string
	// var err error
	const cacheKey = "wx_access_token"
	token, err := session.GetCache(cacheKey)
	if err != nil && "redis: nil" != err.Error() {
		flog.LogFile.Println(err)
	}
	if "" == token {
		url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential"
		url += "&appid=" + config.ConfigVal.WxAppid + "&secret=" + config.ConfigVal.WxSecret
		rsp, err := http.Get(url)
		if err != nil {
			flog.LogFile.Println(err)
		}
		rspBodyj, err := ioutil.ReadAll(rsp.Body)
		rsp.Body.Close()
		if err != nil {
			flog.LogFile.Println(err)
		}
		errRsp := &errorRsp{}
		if err := json.Unmarshal(rspBodyj, errRsp); err != nil {
			flog.LogFile.Println(err)
		}
		if 0 == errRsp.Errcode {
			rspBody := &accessTokenRsp{}
			if err := json.Unmarshal(rspBodyj, rspBody); err != nil {
				flog.LogFile.Println(err)
			}
			token = rspBody.AccessToken
			if err := session.PutCache(cacheKey, token, rspBody.ExpiresIn); err != nil {
				flog.LogFile.Println(err)
			}
		} else {
			flog.LogFile.Println("weixin access token err: ", errRsp.Errmsg)
		}
	}
	return token
}
func JsTicket() string {
	const cacheKey = "wx_jsapi_ticket"
	ticket, err := session.GetCache(cacheKey)
	if err != nil && "redis: nil" != err.Error() {
		flog.LogFile.Println(err)
	}
	if "" == ticket {
		url := "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=" + AccessToken() + "&type=jsapi"
		rsp, err := http.Get(url)
		if err != nil {
			flog.LogFile.Println(err)
		}
		rspBodyj, err := ioutil.ReadAll(rsp.Body)
		rsp.Body.Close()
		if err != nil {
			flog.LogFile.Println(err)
		}
		rspBody := &jsapiTicketRsp{}
		if err := json.Unmarshal(rspBodyj, rspBody); err != nil {
			flog.LogFile.Println(err)
		}
		if 0 == rspBody.Errcode {
			ticket = rspBody.Ticket
			if err := session.PutCache(cacheKey, ticket, rspBody.ExpiresIn); err != nil {
				flog.LogFile.Println(err)
			}
		} else {
			flog.LogFile.Println("weixin jsapi ticket err: ", rspBody.Errmsg)
		}
	}
	return ticket
}
