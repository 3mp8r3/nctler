package utils

import (
	"github.com/go-redis/redis"
)

func GetRedisClient() (*redis.Client, error) {
	addr, passwd, db := conf.Getv("REDIS_ADDR"), conf.Getv("REDIS_PASSWORD"), conf.GetInt("REDIS_DB")
	client := getClient(addr, passwd, db)
	_, err := client.Ping().Result()
	return client, err
}

func getClient(addr, passwd string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})
	return client
}
