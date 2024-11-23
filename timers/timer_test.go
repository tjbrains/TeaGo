package timers_test

import (
	"github.com/tjbrains/TeaGo/logs"
	"github.com/tjbrains/TeaGo/timers"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	t.Log(time.Now(), "start")
	timers.Delay(3*time.Second, func(timer *time.Timer) {
		t.Log(time.Now(), "run task")
	})

	time.Sleep(time.Second * 10)
}

func TestAt(t *testing.T) {
	t.Log(time.Now(), "start")
	timers.At(time.Now().Add(5*time.Second), func(timer *time.Timer) {
		t.Log(time.Now(), "run task")
	})
	//timer.Stop()

	timers.At(time.Now().Add(-5*time.Second), func(timer *time.Timer) {
		t.Log(time.Now(), "run task2")
	})

	time.Sleep(time.Second * 10)
}

func TestEvery(t *testing.T) {
	t.Log(time.Now(), "start")
	i := 0
	var ticker *time.Ticker
	ticker = timers.Every(3*time.Second, func(timer *time.Ticker) {
		t.Log(time.Now(), "run task")
		i++

		if i == 2 {
			ticker.Stop()
		}
	})

	time.Sleep(time.Second * 10)
}

func TestLoop(t *testing.T) {
	var looper = timers.Loop(1*time.Second, func(looper *timers.Looper) {
		logs.Println(time.Now())
	})

	timers.Delay(5*time.Second, func(timer *time.Timer) {
		looper.Stop()
	})

	looper.Wait()
	t.Log("finished")
}

func TestLoop2(t *testing.T) {
	var fromTime = time.Now()
	var looper = timers.Loop(2*time.Second, func(looper *timers.Looper) {
		logs.Println(time.Now())

		if time.Since(fromTime) > 3*time.Second {
			looper.Stop()
			logs.Println("stop")
		}
	})

	looper.Wait()
	t.Log("finished")
}
