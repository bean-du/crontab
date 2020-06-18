package main

import (
	"crontab/worker"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var (
	ConfigFile string
)

func InitCPU() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func InitArgs() {
	flag.StringVar(&ConfigFile, "config", "./worker.json", "指定配置文件./worker.json")
	flag.Parse()
}

func main() {
	var (
		err error
	)
	// 初始化线程
	InitCPU()
	// 初始化命令行参数
	InitArgs()

	// 初始化配置文件
	if err = worker.InitConfig(ConfigFile); err != nil {
		goto ERR
	}
	// 启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	// 启动执行器
	if err = worker.InitExecute(); err != nil  {
		goto ERR
	}
	//初始化任务管理器
	if err = worker.InitJobManager(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Second * 2)
	}
ERR:
	fmt.Println(err)
}
