package controllers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"node/common"
	"node/utils"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/go-redis/redis"
	"golang.org/x/net/context"
)

var (
	conf           *common.Settings
	chanLength     int
	dockerEndpoint string
	dockerVersion  string
)

const (
	// RedisContainerPrefix prefix of redis key
	RedisContainerPrefix = "container:"
	// RedisKeyExpire expire time of redis key when container exited
	RedisKeyExpire = 7 * 24 * 3600
)

var (
	eventOptions = map[string]string{
		// value: filter
		events.ContainerEventType: "type",
		"start":                   "event",
		"stop":                    "event",
		"kill":                    "event",
	}
)

func init() {
	conf = common.GetSettings()
	chanLength = conf.GetInt("CHAN_LENGTH")
	dockerEndpoint = conf.Getv("DOCKER_ENDPOINT")
	dockerVersion = conf.Getv("DOCKER_VERSION")
}

// DockerEventController handle docker event
type DockerEventController struct {
	exit chan struct{}
}

// NewDockerEventController create DockerEventController
func NewDockerEventController() *DockerEventController {
	return &DockerEventController{exit: make(chan struct{})}
}

// Start start controller
func (c *DockerEventController) Start() error {
	logrus.Infof("DockerEventController started...")
	err := c.handleContainerEvent()
	return err
}

// GetControllerName get name of controller
func (c *DockerEventController) GetControllerName() string {
	return "dockerEventHandler"
}

func (c *DockerEventController) handleContainerEvent() error {
	eventChan := make(chan events.Message, chanLength)
	redisCli, err := utils.GetRedisClient()
	if err != nil {
		logrus.Errorf("New Redis Client error: %v", err)
		return errors.New("dockerEventHandler connect to redis error")
	}
	dockerCli, err := client.NewClient(dockerEndpoint, dockerVersion, nil, nil)
	if err != nil {
		logrus.Errorf("New Docker Client error: %v", err)
		return errors.New("dockerEventHandler create docker client error")

	}
	go c.writeEventChan(dockerCli, eventChan)
	for {
		select {
		case event := <-eventChan:
			logrus.Infof("start handle event: %s", event)
			go c.dealEvent(redisCli, dockerCli, event)
		case <-c.exit:
			logrus.Errorf("DockerEventController exited.")
			return errors.New("controller receives stop signal")
		}
	}
	return nil
}

func (c *DockerEventController) dealEvent(redisCli *redis.Client, cli *client.Client, event events.Message) error {
	logrus.Infof("Dealing docker  %s event.", event.Action)
	cJSON, err := cli.ContainerInspect(context.Background(), event.ID)
	if err != nil {
		logrus.Errorf("inspect container error: %v", err)
		return err
	}
	compactCID := cJSON.ID[:12]
	logrus.Infof("container_json: %v", cJSON)
	if event.Action == "start" {
		logrus.Infof("container %s started", cJSON.ID)
		info, err := containerInfoForStart(cli, cJSON)
		if info != nil {
			err = redisCli.HMSet(RedisContainerPrefix+compactCID, info).Err()
			logrus.Infof("insert into redis: key: %s, val: %s, err: %v",
				RedisContainerPrefix+compactCID, info, err)
		}
	} else if event.Action == "stop" || event.Action == "kill" {
		logrus.Infof("container %s %sed", cJSON.ID, event.Action)
		res, err := redisCli.Expire(RedisContainerPrefix+compactCID, time.Duration(RedisKeyExpire)).Result()
		// res, err := redisCli.Del(RedisContainerPrefix + compactCID).Result()
		logrus.Infof("delete redis key: [%s], res: %v, err: %v", RedisContainerPrefix+compactCID, res, err)
	}
	return err
}

func (c *DockerEventController) writeEventChan(cli *client.Client, eventChan chan events.Message) {
	logrus.Infof("-------------------writing event nessage to channel------------------------")
	args := filters.NewArgs()
	for name, value := range eventOptions {
		args.Add(value, name)
	}
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
		case <-c.exit:
			logrus.Errorf("writeEventChan exited")
			return
		}
	}

}

func containerInfoForStart(cli *client.Client, cJSON types.ContainerJSON) (map[string]interface{}, error) {
	containerid := cJSON.ID
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	config := cJSON.Config

	env := make(map[string]string)
	for _, e := range config.Env {
		eArray := strings.Split(e, "=")
		env[eArray[0]] = eArray[1]
	}

	mounts := make(map[string]string)
	for _, m := range cJSON.Mounts {
		mounts[m.Source] = mounts[m.Destination]
	}
	logrus.Infof("env: %v, mount: %v", env, mounts)

	execInfo := ""
	// get info for 3 times
	for i := 1; i <= 3; i++ {
		if execInfo == "" {
			logrus.Infof("docker exec %d times", i)
			output, _ := utils.DockerExec(cli, containerid,
				[]string{"ls"})
			if err != nil {
				logrus.Errorf("docker exec error: %v", err)
				continue
			}
			execInfo = string(output)
			time.Sleep(2 * time.Second)
		}
	}
	execInfo = strings.TrimSpace(execInfo)

	logrus.Infof("execInfo: %s", execInfo)

	outaddr := strings.Trim(conf.Getv("REDIS_ADDR"), "http://")
	host, err := utils.GetIPAddr(outaddr)
	logrus.Infof("-----host: %s", host)
	if err != nil {
		logrus.Errorf("Get host ip error: %v", err)
	}

	info := make(map[string]interface{})
	info["execInfo"] = execInfo
	info["containerid"] = containerid[:12]
	info["host"] = fmt.Sprintf("%s", host)
	info["hostname"] = hostname
	info["time"] = time.Now().String()
	logrus.Infof("app version info: %v", info)
	return info, nil
}

// Stop stop controller
func (c *DockerEventController) Stop() {
	close(c.exit)
}
