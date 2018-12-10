package handler

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

const (
	HTTP_METHOD_GET = "GET"
)

type Resource struct {
}

type Response struct {
	Status string `json:"status"`
}

func (r Resource) Register(router *mux.Router) {

	router.HandleFunc("/health", r.HealthHandler).Methods(HTTP_METHOD_GET)
}

func (r Resource) AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
}

func (r Resource) HealthHandler(writer http.ResponseWriter, request *http.Request) {

	resp := Response{Status: "ok"}
	respbytes, _ := json.Marshal(resp)
	writer.Write(respbytes)
}
