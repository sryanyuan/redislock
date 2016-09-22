package redislock

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
)

var (
	ErrSingleLockInvalidRedisConn = errors.New("SingleLock : Invalid redis conn")
	ErrSingleLockOperationFailed  = errors.New("SingleLock : Operation failed")
	ErrSingleLockNotLocked        = errors.New("SingleLock : Not locked")
	ErrSingleLockInvalidLockValue = errors.New("SingleLock : Invalid lock value")
	ErrSingleLockLockIsUnlocked   = errors.New("SingleLock : Lock is unlocked")
)

var (
	gSingleRedisPool *redis.Pool
)

// SingleLocker is a locker by using single redis node
type SingleLocker struct {
	lockKey   string
	lockValue string
}

// RedisConf is a config struct to create redis connection
type RedisConf struct {
	Address     string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
}

// NewSingleLocker return a SingleLock instance
func NewSingleLocker() *SingleLocker {
	return &SingleLocker{}
}

// Initialize the redis pool
func InitSingleLocker(cfg *RedisConf) {
	if nil != gSingleRedisPool {
		return
	}

	gSingleRedisPool = &redis.Pool{
		MaxIdle:     cfg.MaxIdle,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: time.Duration(cfg.IdleTimeout) * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.Address)
			if err != nil {
				return nil, err
			}
			//	Auth ?
			if len(cfg.Password) != 0 {
				if _, err := c.Do("AUTH", cfg.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
	}
}

func (s *SingleLocker) Lock(keyName string, timeout int64) error {
	conn := gSingleRedisPool.Get()
	if nil == conn {
		return ErrSingleLockInvalidRedisConn
	}
	defer conn.Close()

	//	to generate the unique key
	u, err := uuid.NewUUID()
	if nil != err {
		return err
	}
	s.lockValue = u.String()
	s.lockKey = keyName

	//	try to lock
	rpl, err := redis.String(conn.Do("SET", keyName, s.lockValue, "NX", "PX", timeout))
	if nil != err {
		return err
	}
	if rpl != "OK" {
		return ErrSingleLockOperationFailed
	}

	return nil
}

func (s *SingleLocker) Unlock() error {
	if len(s.lockKey) == 0 ||
		len(s.lockValue) == 0 {
		return ErrSingleLockNotLocked
	}

	conn := gSingleRedisPool.Get()
	if nil == conn {
		return ErrSingleLockInvalidRedisConn
	}
	defer conn.Close()

	//	try to unlock
	//	avoid to unlock a lock not belongs to the locker
	lockValue, err := redis.String(conn.Do("GET", s.lockKey))
	if nil != err {
		return err
	}
	if lockValue != s.lockValue {
		return ErrSingleLockInvalidLockValue
	}

	rpl, err := redis.Int(conn.Do("DEL", s.lockKey))
	if nil != err {
		return err
	}

	if rpl != 1 {
		return ErrSingleLockLockIsUnlocked
	}

	return nil
}
