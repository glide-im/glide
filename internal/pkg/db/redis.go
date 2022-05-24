package db

import (
	"fmt"
	"github.com/go-redis/redis"
)

var Redis *redis.Client

type RedisConfig struct {
	Host        string
	Port        int
	Password    string
	PoolSize    int
	MinIdleConn int
	Db          int
}

func initRedis(c *RedisConfig) {

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	opt := &redis.Options{
		Addr:     addr,
		Password: c.Password,
		DB:       c.Db,
	}
	if c.MinIdleConn > 0 {
		opt.MinIdleConns = c.MinIdleConn
	}
	if c.PoolSize > 0 {
		opt.PoolSize = c.PoolSize
	}
	Redis = redis.NewClient(opt)
}
