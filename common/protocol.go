package common

import (
	"context"
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
// 任务计划
type JobSchedulePlan struct {
	Job *Job
	Expr *cronexpr.Expression
	NextTime time.Time
}
// 任务执行信息
type JobExecuteInfo struct {
	Job *Job
	PlanTime time.Time
	RealTime time.Time
	Ctx context.Context
	CancelFunc context.CancelFunc
}
// 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output []byte
	Err error
	StartTime time.Time
	EndTime time.Time
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

func BuildExecuteInfo(jobPlan *JobSchedulePlan) (jbExecuteInfo *JobExecuteInfo) {

	jbExecuteInfo = &JobExecuteInfo{
		Job: jobPlan.Job,
		PlanTime: jobPlan.NextTime,
		RealTime: time.Now(),
	}
	jbExecuteInfo.Ctx, jbExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
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
// 反序列化 json 数据到 Job 里
func UnPack(data []byte) (j *Job, err error) {
	r := new(Job)
	if err = json.Unmarshal(data, r); err != nil {
		return
	}
	j = r
	return
}
// Extract job name
func ExtractJobName(s string) string {
	return strings.TrimPrefix(s, JOB_SAVE_DIR)
}
func ExtractKillerName(s string) string {
	return strings.TrimPrefix(s, JOB_KILL_DIR)
}