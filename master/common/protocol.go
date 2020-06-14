package common

import "encoding/json"

// 定时任务
type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cron_expr"`
}

type Response struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func BuildResponse(code int, msg string, data interface{})(res []byte, err error)  {
	r := &Response{
		Code: code,
		Msg: msg,
		Data: data,
	}
	return json.Marshal(r)
}