package db

import (
	"gopkg.in/redis.v3"
	"strconv"
	"time"
)

type HomeRedis struct {
	*redis.Client
	LastUseTime		time.Time
}

var RedisMap map[string]*HomeRedis

func NewRedis(addr string, password string, number int) (*HomeRedis, error) {
	if RedisMap == nil {
		RedisMap = make(map[string]*HomeRedis)
	}
	key := addr + strconv.Itoa(number)
	if c, ok := RedisMap[key]; ok {
		c.LastUseTime = time.Now()
		return c, nil
	}
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       int64(number),  // use default DB
	})
	homeRedis := &HomeRedis{r, time.Now()}
	RedisMap[key] = homeRedis
	_, err := homeRedis.Ping().Result()
	return homeRedis, err
}
