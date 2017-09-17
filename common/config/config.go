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
	Project []projectConfig
	QiniuAK  string
	QiniuSK  string
	WxAppid  string
	WxSecret string
}
type projectConfig struct {
	Name string
	Tcp string
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

func GetProjectConfig(name string) (pCfg projectConfig) {
	for _, pc := range ConfigVal.Project {
		if name == pc.Name {
			pCfg = pc
			break
		}
	}
	return
}
