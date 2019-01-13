package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"node/common"
	"node/controllers"
	"node/handler"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	// GitCommit git commit id
	GitCommit = "Unknown"
	// BuildTime build time
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

	cm := controllers.ControllerManager{}
	controller := controllers.NewDockerEventController()
	controllers := make([]common.Controller, 0)
	controllers = append(controllers, controller)
	cm.RegisterAll(controllers)
	cm.StartAll()

	//start http server for health check and pprof
	conf := common.GetSettings()
	pprofPort := conf.Getv("SERVICE_PORT")
	r := handler.Resource{}
	router := mux.NewRouter()
	r.Register(router)
	r.AttachProfiler(router)
	fmt.Println(http.ListenAndServe(":"+pprofPort, router))
	go func() {
		fmt.Println(http.ListenAndServe(":"+pprofPort, router))
	}()

	Run()
}

// Run run process
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
