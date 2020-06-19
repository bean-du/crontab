package worker

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type Configure struct {
	// etcd 配置
	Endpoints []string `json:"endpoints"`
	DialTimeout int `json:"dial_timeout"`
	// MongoDB 配置
	Host string `json:"host"`
	Port int `json:"port"`
	DB string `json:"db"`
	UserName string `json:"user_name"`
	Pwd string `json:"pwd"`
	JobLogBatchSize int `json:"job_log_batch_size"`
	LogCommitTimeout int `json:"log_commit_timeout"`
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
