package main

import (
	"flag"
	"fmt"
	//"net/http"
	//_ "net/http/pprof"
	"os"
	"os/signal"

	"node/common"
	//"node/handler"
	"node/plugins"

	"github.com/Sirupsen/logrus"
	//"github.com/gorilla/mux"
)

var (
	GitCommit = "Unknown"
	BuildTime = "Unknown"
)

func main() {

	version := flag.Bool("v", false, "version info")
	flag.Parse()

	if *version {
		fmt.Println("Git Commit: " + GitCommit)
		fmt.Println("Build Time: " + BuildTime)
		return
	}

	defer func() {
		logrus.Infof("recovering...")
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	logrus.Infof("start listening...")
	pm := plugins.PluginManager{}
	plugin := plugins.EventHandler{}
	plugins := make([]common.BasePlugin, 0)
	plugins = append(plugins, &plugin)
	pm.RegPlugins(plugins)

	start http server for health check and pprof
		conf := common.GetSettings()
		pprof_port := conf.Getv("SERVICE_PORT")
		r := handler.Resource{}
		router := mux.NewRouter()
		r.Register(router)
		r.AttachProfiler(router)
		fmt.Println(http.ListenAndServe(":"+pprof_port, router))
		go func() {
			fmt.Println(http.ListenAndServe(":"+pprof_port, router))
		}()

	Run()
}

func Run() {
	exit := make(chan os.Signal, 0)
	signal.Notify(exit, os.Kill, os.Interrupt)
	for {
		select {
		case <-exit:
			logrus.Errorf("main function exited.")
			return
		}
	}
}
