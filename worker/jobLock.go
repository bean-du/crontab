package worker

import (
	"context"
	"crontab/common"
	"github.com/coreos/etcd/clientv3"
)

// 分布式锁（TXN事物）
type JobLock struct {
	kv      clientv3.KV
	lease   clientv3.Lease
	jobName string // 任务名
	// 用于终止自动续租，实现unlock
	cancelFunc context.CancelFunc
	leaseID clientv3.LeaseID
	isLocked bool
}
// 初始化一把锁
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv: kv,
		lease: lease,
		jobName:jobName,
	}
	return
}

func (jl *JobLock)TryLock() (err error)  {
	var (
		grant *clientv3.LeaseGrantResponse
		leaseID clientv3.LeaseID
		ctx context.Context
		cancelFunc context.CancelFunc
		aliveChan <-chan *clientv3.LeaseKeepAliveResponse
		txn clientv3.Txn
		lockKey string
		txnResp *clientv3.TxnResponse
	)
	// 创建租约
	grant, err = jl.lease.Grant(context.TODO(), 5)
	if err != nil {
		goto FAIL
	}
	leaseID = grant.ID
	// 创建cancelFunc
	ctx, cancelFunc = context.WithCancel(context.TODO())
	// 自动续租
	if aliveChan, err = jl.lease.KeepAlive(ctx, leaseID); err != nil {
		goto FAIL
	}
	// 处理续租应答的协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-aliveChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()
	// 创建 TXN 事物
	txn = jl.kv.Txn(context.TODO())
	// 锁路径
	lockKey = common.JOB_LOCK_DIR + jl.jobName
	// 事物抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey),"=", 0)).
		Then(clientv3.OpPut(lockKey,"",clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockKey))
	if txnResp, err = txn.Commit();err != nil {
		goto FAIL
	}
	if !txnResp.Succeeded {
		err = common.ErrLockOccupied
		goto FAIL
	}
	jl.cancelFunc = cancelFunc
	jl.leaseID = leaseID
	jl.isLocked = true
	// 成功返回锁
	return
	//，失败回滚，释放租约
FAIL:
	cancelFunc()
	jl.lease.Revoke(context.TODO(),leaseID)
	return
}

func (jl *JobLock)Unlock()  {
	if jl.isLocked {
		jl.cancelFunc()
		jl.lease.Revoke(context.TODO(), jl.leaseID)
	}
}