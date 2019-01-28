package common

import (
	"os"
	"strconv"
	"sync"
)

func env(key, defaultVal string) string {
	val, exists := os.LookupEnv(key)
	if exists {
		return val
	}
	return defaultVal
}

// Settings a map of cpnfig
type Settings struct {
	config map[string]string
}

var s *Settings
var once sync.Once

// GetSettings get setting params
func GetSettings() *Settings {
	once.Do(func() {
		s = &Settings{}
		config := make(map[string]string)
		//settings
		config["INFLUXDB_HOST"] = env("INFLUXDB_HOST", "http://127.0.0.1")
		config["INFLUXDB_WRITE_PORT"] = env("INFLUXDB_WRITE_PORT", "8086")
		config["INFLUXDB_READ_PORT"] = env("INFLUXDB_READ_PORT", "8086")
		config["INFLUXDB_DATABASE"] = env("INFLUXDB_DATABASE", "dcos")
		// config["DOCKER_ENDPOINT"] = env("DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
		config["DOCKER_ENDPOINT"] = env("DOCKER_ENDPOINT", "tcp://192.168.56.102:2375")
		config["DOCKER_VERSION"] = env("DOCKER_API_VERSION", "v1.29")
		config["SERVICE_PORT"] = env("SERVICE_PORT", "9999")
		config["REDIS_ADDR"] = env("REDIS_ADDR", "192.168.56.102:6379")
		config["REDIS_DB"] = env("REDIS_DB", "0")
		config["REDIS_PASSWORD"] = env("REDIS_PASSWORD", "")
		config["CHAN_LENGTH"] = env("CHAN_LENGTH", "500")
		s.config = config

	})
	return s
}

// Get get env value by key
func (s *Settings) Get(key string) (string, bool) {
	val, exists := s.config[key]
	return val, exists
}

// Getv get env value by key without check exist
func (s *Settings) Getv(key string) string {
	return s.config[key]
}

// GetInt get int value by key
func (s *Settings) GetInt(key string) int {
	strv := s.config[key]
	intv, _ := strconv.Atoi(strv)
	return intv
}
