package main

import (
	"crontab/master"
	"flag"
	"fmt"
	"runtime"
	"time"
)
var (
	ConfigFile string
)
func InitCPU()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func InitArgs()  {
	flag.StringVar(&ConfigFile, "config","./master.json","指定配置文件./master.json")
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
	if err = master.InitConfig(ConfigFile); err != nil {
		goto ERR
	}
	// 初始化任务管理器
	if err = master.InitJobManager(); err != nil {
		goto ERR
	}
	// init http server thread
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
	for  {
		time.Sleep(time.Second * 2)
	}
ERR:
	fmt.Println(err)
}
