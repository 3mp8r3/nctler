package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/maoqide/nctler/common"
	"github.com/maoqide/nctler/controllers"
	_ "github.com/maoqide/nctler/controllers/docker"
	"github.com/maoqide/nctler/handler"
	"github.com/maoqide/nctler/utils"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// GitCommit git commit id
	GitCommit = "Unknown"
	// BuildTime build time
	BuildTime = "Unknown"
	// Version v1.0
	Version = "v1.0"
	conf    = common.GetSettings()
)

func newNctlerCommand() *cobra.Command {
	opts := &common.NctlerOptions{}
	var flags *pflag.FlagSet
	var cmd = &cobra.Command{
		Use:   "nctler",
		Short: "node controller to do monitoring and so on...",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Version {
				printVersion()
			}
			go ServeHTTP(conf.Getv("SERVICE_PORT"))
			cm := controllers.DefaultControllerManager()
			Run(cm)
			utils.Wait(cm.StopAll)
			return nil

		},
	}
	flags = cmd.Flags()
	flags.BoolVar(&opts.Version, "version", false, "Print version information and quit")

	return cmd
}

func main() {
	cmd := newNctlerCommand()
	// cmd.Flags()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println("Version: " + Version)
	fmt.Println("Git Commit: " + GitCommit)
	fmt.Println("Build Time: " + BuildTime)
	os.Exit(0)
}

// Run start controllers and server
func Run(cm *controllers.ControllerManager) {
	cm.StartAll()
}

// ServeHTTP start http server for health check and pprof
func ServeHTTP(pprofPort string) {
	fmt.Printf("server http on \":%s\"\n", pprofPort)
	r := handler.Resource{}
	router := mux.NewRouter()
	r.Register(router)
	r.AttachProfiler(router)
	fmt.Println(http.ListenAndServe(":"+pprofPort, router))
}
