package master

import (
	"crontab/master/common"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)
// http task interface
type ApiServer struct {
	httpServer *http.Server
}

var (
	//单例对象
	G_ApiServer *ApiServer
)
// 处理 cron 保存
// POST json {"name":"job1", "command":"echo  hello~", "cron_expr":"*/5 * * * * * *"}
func handleSave(w http.ResponseWriter, r *http.Request) {
	// 任务保存在etcd
	var (
		job common.Job
		oldJob *common.Job
		err error
		resData []byte
		postJob string
	)
	// 解析POST 表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	postJob = r.PostForm.Get("job")
	fmt.Println(postJob)
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	if oldJob, err = G_JobManager.SaveJob(&job); err != nil {
		goto ERR
	}
	if resData, err = common.BuildResponse(0, "success",oldJob); err == nil {
		w.Write(resData)
	}
	return
ERR:
	if resData, err = common.BuildResponse(-1, err.Error(),""); err == nil {
		w.Write(resData)
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request)  {

}
// 初始化一个http server单例
func InitApiServer() (err error) {
	var (
		mux *http.ServeMux
		listen net.Listener
		httpServer *http.Server
	)
	mux = http.NewServeMux()
	mux.HandleFunc("/cron/save", handleSave)
	mux.HandleFunc("/cron/delete", handleDelete)
	// Get configure
	conf := GetConf()
	// start watch tcp port
	if listen, err = net.Listen("tcp", ":" + strconv.Itoa(conf.Port)); err != nil {
		return
	}
	httpServer = &http.Server{
		ReadHeaderTimeout: time.Duration(conf.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(conf.WriteTimeout) * time.Millisecond,
		Handler: mux,
	}
	G_ApiServer = &ApiServer{
		httpServer: httpServer,
	}
	go G_ApiServer.httpServer.Serve(listen)
	fmt.Println("http server listen on ::",conf.Port)
	return err
}