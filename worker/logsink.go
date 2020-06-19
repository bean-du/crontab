package worker

import (
	"context"
	"crontab/common"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type LogSink struct {
	client         *mongo.Client
	collection     *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_LogSink *LogSink
)

func (s *LogSink) writeLoop() {
	var (
		logs *common.LogBatch
	)
	for {
		select {
		case log := <-s.logChan:
			if logs == nil {
				logs = &common.LogBatch{}
			}
			logs.Logs = append(logs.Logs, log)
			// 设置一个定时器，超时后自动提交保存日志
			timer := time.AfterFunc(time.Duration(G_Config.LogCommitTimeout)*time.Millisecond,
				func(logs *common.LogBatch) func() {
					return func() {
						s.autoCommitChan <- logs
					}
				}(logs))
			// 如果日志数量达到阈值，就提交保存
			if len(logs.Logs) >= G_Config.JobLogBatchSize {
				s.saveLogs(logs)
				logs = nil
				// 当触发到阈值保存日志时，取消定时器
				timer.Stop()
			}

		case timeoutLogs := <-s.autoCommitChan:
			if timeoutLogs != logs {
				continue
			}
			s.saveLogs(timeoutLogs)
			logs = nil
		}
	}
}

// 批量写入日志
func (s *LogSink) saveLogs(logs *common.LogBatch) {
	_, err := s.collection.InsertMany(context.TODO(), logs.Logs)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *LogSink) AppendLog(log *common.JobLog) {
	select {
	case s.logChan <- log:
	default:

	}

}

func InitLogSink() (err error) {
	var (
		client     *mongo.Client
		collection *mongo.Collection
	)
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", G_Config.UserName, G_Config.Pwd, G_Config.Host, G_Config.Port)
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri)); err != nil {
		return
	}
	collection = client.Database(G_Config.DB).Collection("log")
	G_LogSink = &LogSink{
		client:         client,
		collection:     collection,
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	go G_LogSink.writeLoop()
	return
}
