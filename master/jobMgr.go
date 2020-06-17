package master

import (
	"context"
	"crontab/common"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"time"
)

//
type JobManager struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
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
	G_JobManager = &JobManager{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	return
}
// 保存任务到etcd
func (jm *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		put      *clientv3.PutResponse
		jobValue []byte
		oJob     common.Job
	)
	jobKey := common.JOB_SAVE_DIR + job.Name

	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	put, err = jm.Kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return
	}
	if put.PrevKv != nil {
		if err = json.Unmarshal(put.PrevKv.Value, &oJob); err != nil {
			err = nil
			return
		}
		oldJob = &oJob
	}
	return
}
// 删除一个任务
func (jm *JobManager) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		DelResp *clientv3.DeleteResponse
		ojob    common.Job
	)

	jobKey := common.JOB_SAVE_DIR + name
	if DelResp, err = jm.Kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(DelResp.PrevKvs) != 0 {
		if err = json.Unmarshal(DelResp.PrevKvs[0].Value, &ojob); err != nil {
			err = nil
			return
		}

		oldJob = &ojob
	}
	return
}
// 列出所以任务
func (jm *JobManager)ListJobs() (jobs []*common.Job, err error)  {
	var (
		getResponse *clientv3.GetResponse
	)
	if getResponse, err = jm.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, j := range getResponse.Kvs {
		job := new(common.Job)
		if err :=json.Unmarshal(j.Value,job); err != nil {
			err = nil
			continue
		}
		jobs = append(jobs, job)
	}
	return
}

func (jm *JobManager)SaveKillJob(name string) (err error)  {
	var (
		grant *clientv3.LeaseGrantResponse
	)
	killKey := common.JOB_KILL_DIR + name
	if grant, err = jm.Lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	leaseId := grant.ID
	if _, err = jm.Kv.Put(context.TODO(), killKey, "",clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}