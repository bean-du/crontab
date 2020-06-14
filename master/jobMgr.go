package master

import (
	"context"
	"crontab/master/common"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)
//
type JobManager struct {
	Client *clientv3.Client
	Kv clientv3.KV
	Lease clientv3.Lease
}

var (
	G_JobManager *JobManager

)

func InitJobManager()(err error)  {
	var (
		client *clientv3.Client
		config clientv3.Config
	)

	config = clientv3.Config{
		Endpoints: G_Config.Endpoints,
		DialTimeout: time.Duration(G_Config.DialTimeout) *time.Millisecond,
	}
	if client,err = clientv3.New(config); err != nil {
		return
	}
	kv := clientv3.NewKV(client)
	lease:= clientv3.NewLease(client)
	G_JobManager = &JobManager{
		Client: client,
		Kv: kv,
		Lease: lease,
	}
	return
}

func (jm *JobManager)SaveJob(job *common.Job) (oldJob *common.Job, err error)  {
	var (
		put *clientv3.PutResponse
		jobValue []byte
		oJob common.Job
	)
	jobKey := fmt.Sprintf("/cron/jobs/%s",job.Name)

	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	put, err = jm.Kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return
	}
	if put.PrevKv != nil  {
		if err = json.Unmarshal(put.PrevKv.Value, &oJob); err != nil {
			err = nil
			return
		}
		oldJob = &oJob
	}
	return
}