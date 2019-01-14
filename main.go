package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/maoqide/nctler/common"
	"github.com/maoqide/nctler/controllers"
	_ "github.com/maoqide/nctler/controllers/docker"
	"github.com/maoqide/nctler/handler"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	// GitCommit git commit id
	GitCommit = "Unknown"
	// BuildTime build time
	BuildTime = "Unknown"
	version   *bool
	cm        *controllers.ControllerManager
)

func init() {
	version = flag.Bool("v", false, "version info")
	flag.Parse()
}

func main() {

	if *version {
		fmt.Println("Git Commit: " + GitCommit)
		fmt.Println("Build Time: " + BuildTime)
		return
	}
	cm = controllers.DefaultControllerManager()

	defer func() {
		logrus.Infof("main panic...")
		if err := recover(); err != nil {
			logrus.Infof("exit.")
		}
	}()

	// cm := controllers.NewControllerManager()
	// controller := dockerctl.NewEventController()
	// controllers := make([]common.Controller, 0)
	// controllers = append(controllers, controller)
	// cm.RegisterAll(controllers)
	// cm.StartAll()
	cm.StartAll()

	//start http server for health check and pprof
	conf := common.GetSettings()
	pprofPort := conf.Getv("SERVICE_PORT")
	r := handler.Resource{}
	router := mux.NewRouter()
	r.Register(router)
	r.AttachProfiler(router)
	// fmt.Println(http.ListenAndServe(":"+pprofPort, router))
	go func() {
		fmt.Println(http.ListenAndServe(":"+pprofPort, router))
	}()
	Run(cm.StopAll)
}

// Run run process
func Run(f func()) {
	logrus.Infof("running...")
	exit := make(chan os.Signal)
	// signal.Notify(exit, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT)
	signal.Notify(exit, os.Kill, os.Interrupt)
	for {
		select {
		case <-exit:
			logrus.Errorf("main function exited.")
			f()
			return
		}
	}
}
