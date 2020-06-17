package worker

import (
	"context"
	"crontab/common"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

//
type JobManager struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
	Watcher clientv3.Watcher
}

var (
	G_JobManager *JobManager
)

func InitJobManager() (err error) {
	var (
		client *clientv3.Client
		config clientv3.Config
	)

	config = clientv3.Config{
		Endpoints:   G_Config.Endpoints,
		DialTimeout: time.Duration(G_Config.DialTimeout) * time.Millisecond,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)
	G_JobManager = &JobManager{
		Client: client,
		Kv:     kv,
		Lease:  lease,
		Watcher: watcher,
	}
	// 启动监听
	G_JobManager.WatchJobs()
	return
}

func (jm *JobManager)WatchJobs()(err error)  {
	var (
		job *common.Job
		jobEvent *common.Event
	)
	get, err := jm.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
	if err != nil {
		return
	}
	// 当前有哪些任务
	for _, jobJson := range get.Kvs {
		if job, err = common.UnPack(jobJson.Value); err == nil {
			// TODO: 把这个Job同步到scheduler调度协程
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			fmt.Println("push job event to scheduler:", *jobEvent)
			// 将变化推送给scheduler
			G_Scheduler.PushJobEvent(jobEvent)
		}
	}

	// 启动一个监听协程，从该revision后监听变化事件
	go func() {

		// 从get时刻的revision开始监听
		watchStartRevision := get.Header.Revision + 1
		// 启动监听 "/cron/jobs/"的变化
		watch := jm.Watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision),clientv3.WithPrefix())
		// 处理监听
		for watchChan := range watch {
			for _,e := range watchChan.Events {
				switch e.Type {
				case mvccpb.PUT:// 任务保存事件 （Event）
					job, err = common.UnPack(e.Kv.Value)
					if err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
					// 将变化推送给scheduler
					G_Scheduler.PushJobEvent(jobEvent)
					fmt.Println("push job event to scheduler:", *jobEvent)
					//  构造一个event 事件
					// TODO: 反序列化Job，然后推送给调度协程 scheduler
				case mvccpb.DELETE: // 任务被删除了
					// 提取任务名
					jobName := common.ExtractJobName(string(e.Kv.Key))
					// 构造一个删除事件 （Event）
					job := &common.Job{
						Name: jobName,
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
					// TODO: 推送一个删除事件给任务调度 scheduler
					// 将变化推送给scheduler
					G_Scheduler.PushJobEvent(jobEvent)
					fmt.Println("push job event to scheduler:", *jobEvent)
				}
			}
		}

	}()

	return
}