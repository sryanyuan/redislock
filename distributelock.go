package redislock

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	gDistributeRedisPoolMap map[string]*redis.Pool
)

func init() {
	gSingleRedisPoolMap = make(map[string]*redis.Pool)
}

type distributeLocker struct {
	lockKey   string
	lockValue string
	redisPool []*redis.Pool
}

func NewDistributeLocker(cfg []*RedisConf) Locker {
	if nil == cfg {
		return nil
	}

	lk := &distributeLocker{}

	for _, v := range cfg {
		var pool *redis.Pool
		pool, ok := gDistributeRedisPoolMap[v.Address]
		if !ok {
			pool = &redis.Pool{
				MaxIdle:     v.MaxIdle,
				MaxActive:   v.MaxActive,
				IdleTimeout: time.Duration(v.IdleTimeout) * time.Second,
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
				Dial: func() (redis.Conn, error) {
					c, err := redis.Dial("tcp", v.Address)
					if err != nil {
						return nil, err
					}
					//	Auth ?
					if len(v.Password) != 0 {
						if _, err := c.Do("AUTH", v.Password); err != nil {
							c.Close()
							return nil, err
						}
					}
					return c, nil
				},
			}
			gDistributeRedisPoolMap[v.Address] = pool
		}
	}

	return lk
}

func (d *distributeLocker) Lock(lockKey string, timeout int64) error {
	return nil
}

func (d *distributeLocker) Unlock() error {
	return nil
}
