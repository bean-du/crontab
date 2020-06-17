package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

// 定时任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cron_expr"`
}

type JobSchedulePlan struct {
	Job *Job
	Expr *cronexpr.Expression
	NextTime time.Time
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type Event struct {
	EventType int `json:"event_type"`
	Job *Job `json:"job"`
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error){
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &JobSchedulePlan{
		Job: job,
		Expr: expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

func BuildJobEvent(eventType int, job *Job) *Event  {
	return &Event{
		EventType: eventType,
		Job: job,
	}
}

func BuildResponse(code int, msg string, data interface{}) (res []byte, err error) {
	r := &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	return json.Marshal(r)
}

func UnPack(data []byte) (j *Job, err error) {
	r := new(Job)
	if err = json.Unmarshal(data, r); err != nil {
		return
	}
	j = r
	return
}

func ExtractJobName(s string) string {
	return strings.TrimPrefix(s, JOB_SAVE_DIR)
}