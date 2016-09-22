# redislock
redislock is a distribute lock using redis

## how to use

There are 2 locker types in this package in design

### SingleLocker

SingleLocker is a locker use one redis server.Before you get a locker, you need initialize it as below :

	InitSingleLocker(&RedisConf{
		Address: "localhost:6379",
	})

Then you can get a locker

	locker := NewSingleLocker()

You can lock it with the lock key and timeout milliseconds. Every locker has a lock key, and if 2 locker has the same lock key and lock it at the same time, only one locker can lock success before another locker is unlock.

When need lock a locker, you should

	for {
		if err := locker.Lock("db1", 10000); nil != err {
			time.Sleep(time.Millisecond * 10)
			t.Log("Sleeping to wait for lock, err :", err)
		} else {
			// lock success
			break
		}
	}

After you lock a locker, you must unlock it, otherwise others cannot lock the locker until the lock is timeout.

	locker.Unlock()

### DistributeLocker

As design, DistributeLocker uses redis cluster, and if one redis server is down, it will not be affected. But now it is unavailable.