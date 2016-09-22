package redislock

// Locker is an interface that has lock and unlock method
type Locker interface {
	Lock(string, int64) error
	Unlock() error
}
