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
	ErrSingleLockInvalidRedisConf = errors.New("SingleLock : Invalid redis conf")
)

var (
	gSingleRedisPool    *redis.Pool
	gSingleRedisPoolMap map[string]*redis.Pool
)

func init() {
	gSingleRedisPoolMap = make(map[string]*redis.Pool)
}

// SingleLocker is a locker by using single redis node
type singleLocker struct {
	lockKey   string
	lockValue string
	redisPool *redis.Pool
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
func NewSingleLocker(cfg *RedisConf) Locker {
	var pool *redis.Pool
	pool, ok := gSingleRedisPoolMap[cfg.Address]
	if !ok {
		pool = &redis.Pool{
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
		gSingleRedisPoolMap[cfg.Address] = pool
	}

	return &singleLocker{
		redisPool: pool,
	}
}

// Lock locks the locker
func (s *singleLocker) Lock(keyName string, timeout int64) error {
	conn := s.redisPool.Get()
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

// Unlock unlocks the locker
func (s *singleLocker) Unlock() error {
	if len(s.lockKey) == 0 ||
		len(s.lockValue) == 0 {
		return ErrSingleLockNotLocked
	}

	conn := s.redisPool.Get()
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
