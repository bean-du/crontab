package master

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type Configure struct {
	// http API 接口配置
	Port         int `json:"port"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
	// etcd 配置
	Endpoints []string `json:"endpoints"`
	DialTimeout int `json:"dial_timeout"`
}

var (
	G_Config *Configure
	G_once sync.Once
)

func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Configure
	)
	G_once.Do(func() {
		if content, err = ioutil.ReadFile(filename); err != nil {
			return
		}

		if err = json.Unmarshal(content, &conf); err != nil {
			return
		}
	})
	G_Config = &conf
	return
}

func GetConf() *Configure {
	return G_Config
}
