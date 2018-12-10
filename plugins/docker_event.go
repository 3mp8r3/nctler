package plugins

import (
	// "encoding/json"

	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"node/common"
	"node/utils"

	"github.com/Sirupsen/logrus"
	// "github.com/docker/engine-api/client"
	// "github.com/docker/engine-api/types"
	// "github.com/docker/engine-api/types/events"
	// "github.com/docker/engine-api/types/filters"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/go-redis/redis"
	"golang.org/x/net/context"
)

var (
	conf            *common.Settings
	chan_length     int
	docker_endpoint string
	docker_version  string
)

const (
	REDIS_TASKID_PREFIX      = "taskid:"
	REDIS_CONTAINERID_PREFIX = "containerid:"
)

var (
	EVENTS_OPTIONS = map[string]string{
		// value: filter
		events.ContainerEventType: "type",
		"start":                   "event",
		"stop":                    "event",
		"kill":                    "event",
	}
)

func init() {
	conf = common.GetSettings()
	chan_length = conf.GetInt("CHAN_LENGTH")
	docker_endpoint = conf.Getv("DOCKER_ENDPOINT")
	docker_version = conf.Getv("DOCKER_VERSION")
}

type EventHandler struct {
	//conf *common.Settings
}

//type version_info struct {
//	Taskid      string `json:"taskid"`
//	Containerid string `json:"containerid"`
//	Appid       string `json:"appid"`
//	Buildnum    string `json:"buildnum"`
//	Checkpath   string `json:"checkpath"`
//	Checkproto  string `json:"checkproto"`
//	Exists      string `json:"exists"`
//	Host        string `json:"host"`
//	Hostname    string `json:"hostname"`
//	Hostpath    string `json:"hostpath"`
//	Md5         string `json:"md5"`
//}

func (handler *EventHandler) Init() error {
	logrus.Infof("EventHandler init")
	//handler.conf = common.GetSettings()
	handler.handleContainerEvent()
	return nil
}

func (handler *EventHandler) GetPluginName() string {
	return "DockerEventHandler"
}

func (handler *EventHandler) handleContainerEvent() {
	logrus.Infof("------------handleContainerEvent started------------------")
	eventChan := make(chan events.Message, chan_length)
	redis_client, err := utils.GetRedisClient()
	if err != nil {
		logrus.Errorf("New Redis Client error: %v", err)
		panic("connect to redis error")
	}
	docker_cli, err := client.NewClient(docker_endpoint, docker_version, nil, nil)
	if err != nil {
		logrus.Errorf("New Docker Client error: %v", err)
	}
	go handler.writeEventChan(docker_cli, eventChan)
	for {
		select {
		case event := <-eventChan:
			logrus.Infof("start handle event: %s", event)
			go handler.dealEvent(redis_client, docker_cli, event)
		}
	}

}

func (handler *EventHandler) dealEvent(redis_client *redis.Client, cli *client.Client, event events.Message) error {
	logrus.Infof("Dealing docker  %s event.", event.Action)
	c_json, err := cli.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		logrus.Errorf("inspect container error: %v", err)
		return err
	}
	compact_contianerid := c_json.ID[:12]
	logrus.Infof("container_json: %v", c_json)
	if event.Action == "start" {
		logrus.Infof("container %s started", c_json.ID)
		info, err := containerInfoForStart(cli, c_json)
		if info != nil {
			err = redis_client.Set(REDIS_TASKID_PREFIX+info["taskid"].(string), compact_contianerid, 0).Err()
			logrus.Infof("insert into redis: key: %s, val: %s, err: %v",
				REDIS_TASKID_PREFIX+info["taskid"].(string), compact_contianerid, err)
			err = redis_client.HMSet(REDIS_CONTAINERID_PREFIX+compact_contianerid, info).Err()
			logrus.Infof("insert into redis: key: %s, val: %s, err: %v",
				REDIS_CONTAINERID_PREFIX+compact_contianerid, info, err)
		}
	} else if event.Action == "stop" || event.Action == "kill" {
		logrus.Infof("container %s %sed", c_json.ID, event.Action)
		// compact_contianerid := c_json.ID[:12]
		taskid, err := redis_client.HGet(REDIS_CONTAINERID_PREFIX+compact_contianerid, "taskid").Result()
		if err != nil {
			logrus.Warningf("<container %s>, get container taskid err: %v", event.Action, err)
			return err
		}
		res, err := redis_client.Del(REDIS_TASKID_PREFIX+taskid, REDIS_CONTAINERID_PREFIX+compact_contianerid).Result()
		logrus.Infof("delete redis key: [%s, %s], res: %v, err: %v", REDIS_TASKID_PREFIX+taskid, REDIS_CONTAINERID_PREFIX+compact_contianerid, res, err)
	}
	return nil
}

func (handler *EventHandler) writeEventChan(cli *client.Client, eventChan chan events.Message) {
	logrus.Infof("-------------------writing event nessage to channel------------------------")
	args := filters.NewArgs()
	for name, value := range EVENTS_OPTIONS {
		args.Add(value, name)
	}
	// args.Add("event", "start")
	// args.Add("event", "stop")
	// body, err := cli.Events(context.Background(), types.EventsOptions{Filters: args})
	messages, errs := cli.Events(context.Background(), types.EventsOptions{Filters: args})
	for {
		select {
		case err := <-errs:
			if err != nil && err != io.EOF {
				logrus.Errorf("decode event message error: %v", err)
			}
		case event := <-messages:
			logrus.Infof("write docker event message: %v", event)
			eventChan <- event
		}
	}

	// for old version docker api
	// dec := json.NewDecoder(body)
	// for {
	// 	var event events.Message
	// 	err := dec.Decode(&event)
	// 	if err != nil && err == io.EOF {
	// 		logrus.Errorf("decode event message error: %v", err)
	// 		continue
	// 	}
	// 	logrus.Infof("write docker event message: %v", event)
	// 	eventChan <- event
	// }

}

func containerInfoForStart(cli *client.Client, c_json types.ContainerJSON) (map[string]interface{}, error) {
	containerid := c_json.ID
	mounts := c_json.Mounts
	logrus.Infof("mounts: %v", mounts)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	config := c_json.Config
	app_id, taskid, appname := "", "", "app.tar"
	// default checkproto COMMAND
	checkproto, checkpath := "COMMAND", ""
	hostpath, exists, md5, buildnum := "", "0", "", ""

	logrus.Infof("-----------env")
	for _, e := range config.Env {
		e_array := strings.Split(e, "=")
		if e_array[0] == "MARATHON_APP_ID" {
			logrus.Infof("MARATHON_APP_ID: %s", e_array[1])
			app_info := strings.Split(e_array[1], ".")
			app_id = strings.Trim(app_info[0], "/")
		} else if e_array[0] == "PACKAGE_CHECK_PROTO" {
			if e_array[1] == "HTTP" || e_array[1] == "COMMAND" {
				checkproto = e_array[1]
				logrus.Infof("PACKAGE_CHECK_PROTO: %s", checkproto)
			} else {
				checkproto = ""
				logrus.Infof("PACKAGE_CHECK_PROTO: %s", "None")
			}
		} else if e_array[0] == "PACKAGE_CHECK_PATH" {
			checkpath = e_array[1]
		} else if e_array[0] == "MESOS_TASK_ID" {
			taskid = e_array[1]
		} else if e_array[0] == "PKG" {
			appname = e_array[1]
		}
	}

	logrus.Infof("-----mounts")
	if checkproto == "HTTP" {
		//TODO utils.HttpGet
		if checkpath == "" {
			checkpath = "/version"
		}
		body, _ := utils.HttpGet(checkpath)
		buildnum = string(body)
	} else if checkproto == "COMMAND" {
		//TODO utils.DockerExec
		if checkpath == "" {
			checkpath = "/root/" + appname
		}
		// get version for 3 times
		for i := 1; i <= 3; i++ {
			if buildnum == "" {
				logrus.Infof("docker exec %d times", i)
				output, _ := utils.DockerExec(cli, containerid,
					[]string{"tar", "-xzO", "-f", checkpath, "version.txt"})
				if err != nil {
					logrus.Errorf("docker exec error: %v", err)
					continue
				}
				buildnum = string(output)
				time.Sleep(1500 * time.Millisecond)
			}
		}
		buildnum = strings.TrimSpace(buildnum)
		logrus.Infof("buildnum: %s", buildnum)
	}
	for _, m := range mounts {
		if m.Destination == "/root" {
			logrus.Infof("----------------------------")
			hostpath = m.Source
			filepath := hostpath + "/" + appname
			logrus.Infof("filepath: %s", filepath)
			if utils.FileExists(filepath) {
				exists = "1"
				md5, err = utils.MD5File(filepath)
				if err != nil {
					logrus.Errorf("MD5 file error: %v", err)
					md5 = ""
				}
			}
		}
	}
	logrus.Infof("---appid: %s", app_id)
	outaddr := strings.Trim(conf.Getv("REDIS_ADDR"), "http://")
	host, err := utils.GetIpAddr(outaddr)
	logrus.Infof("-----host: %s", host)
	if err != nil {
		logrus.Errorf("Get host ip error: %v", err)
	}

	info := make(map[string]interface{})
	info["taskid"] = taskid
	info["buildnum"] = buildnum
	info["containerid"] = containerid[:12]
	info["appid"] = app_id
	info["checkpath"] = checkpath
	info["checkproto"] = checkproto
	info["exists"] = exists
	info["host"] = fmt.Sprintf("%s", host)
	info["hostname"] = hostname
	info["hostpath"] = hostpath
	info["md5"] = md5
	info["time"] = time.Now().String()
	logrus.Infof("app version info: %v", info)
	return info, nil
}

/*
***************************deprecated****************************
 */
//!deprecated
func (handler *EventHandler) handleStartContainer_0() {
	logrus.Infof("------------HandleStartContainer started------------------")
	endpoint := conf.Getv("DOCKER_ENDPOINT")
	version := conf.Getv("DOCKER_VERSION")
	cli, err := client.NewClient(endpoint, version, nil, nil)
	if err != nil {
		logrus.Errorf("New Docker Client error: %v", err)
	}
	args := filters.NewArgs()
	args.Add("event", "start")
	messages, errs := cli.Events(context.Background(), types.EventsOptions{Filters: args})
	// body, err := cli.Events(context.Background(), types.EventsOptions{Filters: args})
	// if err != nil {
	// 	logrus.Errorf("Listen Docker Events error: %v", err)
	// }

	// dec := json.NewDecoder(body)
	influxdb_cli, err := utils.GetInfluxDBWriteClient()
	for {
		select {
		case err := <-errs:
			if err != nil && err != io.EOF {
				logrus.Errorf("decode event message error: %v", err)
			}
		case event := <-messages:
			logrus.Infof("docker event message: %v", event)
			c_json, err := cli.ContainerInspect(context.Background(), event.ID)
			if err != nil {
				logrus.Errorf("inspect container error: %v", err)
				continue
			}
			logrus.Infof("container_json: ", c_json)
			tags, field, err := containerInfoForStart_0(cli, c_json)
			if tags != nil {
				measurement := "app_version"
				logrus.Infof("influxdb measurement: %s, tags: %v, field: %v", measurement, tags, field)
				ok, err := utils.WriteData(influxdb_cli, measurement, tags, field, time.Now())
				logrus.Infof("influxdb writedata, ok: %v, err: %v", ok, err)
			}
		}
	}
}

//!deprecated
func containerInfoForStart_0(cli *client.Client, c_json types.ContainerJSON) (map[string]string, map[string]interface{}, error) {
	containerid := c_json.ID
	mounts := c_json.Mounts
	logrus.Infof("mounts: %v", mounts)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	config := c_json.Config
	app_id, taskid, appname := "", "", "app.tar"
	checkproto, checkpath := "", ""
	hostpath, exists, md5, buildnum := "", "0", "", ""
	logrus.Infof("-----------env")
	for _, e := range config.Env {
		e_array := strings.Split(e, "=")
		if e_array[0] == "MARATHON_APP_ID" {
			logrus.Infof("MARATHON_APP_ID: %s", e_array[1])
			app_info := strings.Split(e_array[1], ".")
			app_id = strings.Trim(app_info[0], "/")
		} else if e_array[0] == "PACKAGE_CHECK_PROTO" {
			if e_array[1] == "HTTP" || e_array[1] == "COMMAND" {
				checkproto = e_array[1]
				logrus.Infof("PACKAGE_CHECK_PROTO: %s", checkproto)
			} else {
				checkproto = ""
				logrus.Infof("PACKAGE_CHECK_PROTO: %s", "None")
			}
		} else if e_array[0] == "PACKAGE_CHECK_PATH" {
			checkpath = e_array[1]
		} else if e_array[0] == "MESOS_TASK_ID" {
			taskid = e_array[1]
		} else if e_array[0] == "PKG" {
			appname = e_array[1]
		}
	}
	logrus.Infof("-----mounts")
	if checkproto == "HTTP" {
		//TODO utils.HttpGet
		if checkpath == "" {
			checkpath = "/version"
		}
		body, _ := utils.HttpGet(checkpath)
		buildnum = string(body)
	} else if checkproto == "COMMAND" {

		checkpath = "/root" + appname
		output, _ := utils.DockerExec(cli, containerid, []string{"tar", "-xzO", "-f", checkpath, "version.txt"})
		buildnum = string(output)
		buildnum = strings.TrimSpace(buildnum)
		//buildnum = strings.SplitN(buildnum, ":", 2)[1]
		logrus.Infof("buildnum: %s", buildnum)
	}
	for _, m := range mounts {
		if m.Destination == "/root" {
			logrus.Infof("----------------------------")
			hostpath = m.Source
			filepath := hostpath + "/" + appname
			logrus.Infof("filepath: %s", filepath)
			if utils.FileExists(filepath) {
				exists = "1"
				md5, err = utils.MD5File(filepath)
				if err != nil {
					logrus.Errorf("MD5 file error: %v", err)
					md5 = ""
				}
			}
		}
	}
	logrus.Infof("---appid: %s, hostpath: %s", app_id, hostpath)
	if app_id == "" || hostpath == "" {
		return nil, nil, nil
	}
	conf := common.GetSettings()
	outaddr := strings.Trim(conf.Getv("INFLUXDB_HOST")+":"+conf.Getv("INFLUXDB_READ_PORT"), "http://")
	host, err := utils.GetIpAddr(outaddr)
	logrus.Infof("-----host: %s", host)
	if err != nil {
		logrus.Errorf("Get host ip error: ", err)
	}
	tags := map[string]string{"appid": app_id}
	fields := map[string]interface{}{
		"containerid": containerid,
		"host":        fmt.Sprintf("%s", host),
		"hostname":    hostname,
		"md5":         md5,
		"checkproto":  checkproto,
		"checkpath":   checkpath,
		"buildnum":    buildnum,
		"exists":      exists,
		"hostpath":    hostpath,
		"taskid":      taskid,
	}
	logrus.Infof("influxdb: %v, %v", tags, fields)
	return tags, fields, nil
}
