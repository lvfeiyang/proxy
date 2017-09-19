package config

import (
	"encoding/json"
	"github.com/lvfeiyang/proxy/common/flog"
	"io/ioutil"
	"runtime"
)

type config struct {
	RedisUrl string
	MongoUrl string
	HtmlPath string
	Project []ProjectConfig
	QiniuAK  string
	QiniuSK  string
	WxAppid  string
	WxSecret string
}
type ProjectConfig struct {
	Name string
	Tcp string
	Http string
	Proxy bool
}

var ConfigVal = &config{}

func Init() {
	var filePath string
	if "linux" == runtime.GOOS {
		filePath = "/root/workspace/xiaobai/config"
	} else {
		filePath = "C:\\Users\\lxm19\\config" //lxm19
	}
	conf, err := ioutil.ReadFile(filePath)
	if err != nil {
		flog.LogFile.Fatal(err)
	}
	err = json.Unmarshal(conf, ConfigVal)
	if err != nil {
		flog.LogFile.Fatal(err)
	}
}

func GetProjectConfig(name string) (pCfg ProjectConfig) {
	for _, pc := range ConfigVal.Project {
		if name == pc.Name {
			pCfg = pc
			return
		}
	}
	// flog.LogFile.Fatal("no "+name+" project!")
	return
}
