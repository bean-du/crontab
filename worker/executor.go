package worker

import (
	"crontab/common"
	"math/rand"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	G_Executor *Executor
)

func (e *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	// 执行shell命令协程
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
		)
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}
		// 初始化锁
		jobLock = G_JobManager.CreateJobLock(info.Job.Name)
		// 执行shell命令
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
		result.StartTime = time.Now()
		err = jobLock.TryLock()
		// 执行完成之后 释放锁
		defer jobLock.Unlock()
		if  err != nil {
			// 上锁失败
			result.Err = err
			result.EndTime = time.Now()
		}else {
			// 重置任务启动时间
			result.StartTime = time.Now()
			cmd = exec.CommandContext(info.Ctx, "/bin/bash", "-c", info.Job.Command)
			// 执行并捕获输出
			output, err = cmd.CombinedOutput()
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		// 执行完成后，将执行结果返回给 Scheduler， Scheduler 会从 ExecutingTable 中删除执记录
		G_Scheduler.PushJobResult(result)
	}()
}

// 初始化执行器

func InitExecute() (err error) {
	G_Executor = &Executor{}
	return
}
