package worker

import (
	"crontab/common"
	"fmt"
	"time"
)

type Scheduler struct {
	jobEventChan chan *common.Event // 任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan
	jobExecutingTable map[string]*common.JobExecuteInfo
	jobResultChan chan *common.JobExecuteResult
}

var (
	G_Scheduler *Scheduler
)
// 处理任务事件在内存中的同步
func (s *Scheduler)handleJobEvent(jobEvent *common.Event)  {
	var (
		jobSchedulerPlan *common.JobSchedulePlan
		err error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulerPlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		//同步到本地内存，与 etcd 同步
		s.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
	case common.JOB_EVENT_DELETE:
		if _,ok := s.jobPlanTable[jobEvent.Job.Name]; ok {
			delete(s.jobPlanTable,jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL:// 强杀任务事件
	// 取消command执行
		if info, executing := s.jobExecutingTable[jobEvent.Job.Name];executing {
			info.CancelFunc()
		}
	}
}
// 重新计算任务调度状态
func (s *Scheduler)TryScheduler() (d time.Duration)  {

	if len(s.jobPlanTable) == 0 {
		time.Sleep(1 * time.Second)
		return
	}
	var (
		jobPlan *common.JobSchedulePlan
		nearTime *time.Time
	)
	// 1. 遍历所有任务
	now := time.Now()
	for _,jobPlan = range s.jobPlanTable {
		// 如果它早与当前时间，或者等于当前时间
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// 尝试执行任务
			s.TryStartJbo(jobPlan)
			// 重置任务的下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 统计最近一个要过期的任务事件
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime){
			nearTime = &jobPlan.NextTime
		}

	}
	// 3. 统计最近要过期的任务时间 （ N秒后过期， N == d）
	// 下次调度睡眠时间（最近要执行任务的时间 - 当前时间）
	d = (*nearTime).Sub(now)
	return
}
// 调度协程
func (s *Scheduler)SchedulerLoop()  {
	var (
		jobEvent *common.Event
		scheduleAfter time.Duration
	)
	// 初始化一次 计算睡眠时间
	scheduleAfter = s.TryScheduler()
	// 调度的延迟定时器
	scheduleTimer := time.NewTimer(scheduleAfter)
	// 定时任务
	for {
		select {
		case jobEvent = <-s.jobEventChan: // 监听任务变化事件
			s.handleJobEvent(jobEvent)
		case <-scheduleTimer.C:
			// 最近的任务到期了
		case jobResult := <-s.jobResultChan: // 监听任务执行结果
			s.handleJobResult(jobResult)
		}
		scheduleAfter = s.TryScheduler()
		// 重置定时器
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 推送任务变化事件
func (s *Scheduler)PushJobEvent(jobEvent *common.Event)  {
	s.jobEventChan <- jobEvent
}

func (s *Scheduler) TryStartJbo(jobPlan *common.JobSchedulePlan) {
	// 执行任务可能会很久，1分钟调度60次，但只能执行一次，防止并发
	if _, ok := s.jobExecutingTable[jobPlan.Job.Name]; ok {
		fmt.Printf("正在执行中，跳过任务名为[%s]的执行\n", jobPlan.Job.Name)
		return
	}
	// 构建执行状态
	info := common.BuildExecuteInfo(jobPlan)
	// 保存执行状态
	s.jobExecutingTable[jobPlan.Job.Name] = info
	// 执行命令
	G_Executor.ExecuteJob(info)
	//fmt.Printf("执行命令:%s , 计划执行时间：%v， 真实执行时间： %v \n", jobPlan.Job.Name, info.PlanTime, info.RealTime)
}

func (s *Scheduler)PushJobResult(res *common.JobExecuteResult)  {
	s.jobResultChan <- res
}

func (s *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	if _, ok := s.jobExecutingTable[result.ExecuteInfo.Job.Name]; ok {
		delete(s.jobExecutingTable,result.ExecuteInfo.Job.Name)
	}
	fmt.Printf("[%s]任务执行完成! , 任务执行结果：%s，执行时间： %v，错误信息： %v \n", result.ExecuteInfo.Job.Name,result.Output,result.EndTime.Sub(result.StartTime),result.Err)
}
// 初始化调度器
func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan: make(chan *common.Event, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan: make(chan *common.JobExecuteResult, 1000),
	}
	//启动调度协程
	go G_Scheduler.SchedulerLoop()
	return
}