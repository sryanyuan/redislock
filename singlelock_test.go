package redislock

import (
	"sync"
	"testing"
	"time"
)

func TestSingleLock_Lock(t *testing.T) {
	InitSingleLocker(&RedisConf{
		Address: "localhost:6379",
	})

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(routineId int) {
			defer wg.Done()
			t.Log("Routine ", routineId, "start")
			locker := NewSingleLocker()

			for {
				if err := locker.Lock("db1", 10000); nil != err {
					time.Sleep(time.Millisecond * 10)
					t.Log("Sleeping to wait for lock, err : ", err)
				} else {
					break
				}
			}

			t.Log("Get lock : ", routineId)
			err := locker.Unlock()
			if nil != err {
				t.Error("Lock of routine ", routineId, "release failed : ", err)
				t.FailNow()
			} else {
				t.Log("Lock of routine ", routineId, "released success")
			}
		}(i)
	}

	wg.Wait()
}
