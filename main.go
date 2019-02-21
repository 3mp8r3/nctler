package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/maoqide/nctler/common"
	"github.com/maoqide/nctler/controllers"
	_ "github.com/maoqide/nctler/controllers/docker"
	"github.com/maoqide/nctler/handler"
	"github.com/maoqide/nctler/utils"

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
	go func() {
		fmt.Println(http.ListenAndServe(":"+pprofPort, router))
	}()
	utils.Wait(cm.StopAll)
}
