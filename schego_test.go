package schego

import (
	"sync"
	"testing"
	"time"

	"github.com/mikespook/golib/autoinc"
)

var (
	ai = autoinc.New(0, 1)
)

type _task struct {
	t  *testing.T
	wg *sync.WaitGroup
}

func (t *_task) Exec(id interface{}) error {
	t.t.Logf("%s Shoot!", id)
	t.wg.Done()
	return nil
}
func (t *_task) Cancel(id interface{}) error {
	t.t.Log("%s Cancel!", id)
	t.wg.Done()
	return nil
}

func TestServe(t *testing.T) {
	sche := New(time.Second)
	sche.ErrorHandler = func(evt Event, err error) {
		t.Errorf("%d, %s", evt.Id, err)
	}
	go sche.Serve()
	n := time.Duration(time.Now().UnixNano())

	var wg sync.WaitGroup
	wg.Add(6)
	id := ai.Id()
	sche.Add(Event{
		Id:       id,
		Start:    n + time.Second,
		Interval: time.Second,
		// negative number means running for ever;
		// 0 or 1 is one time event;
		// other numbers show how many time will be executed.
		Iterate: 1,
		Task:    &_task{t, &wg},
	})
	id = ai.Id()
	sche.Add(Event{
		Id:       id,
		Start:    n + time.Second,
		Interval: time.Second,
		Iterate:  3,
		Task:     &_task{t, &wg},
	})
	id = ai.Id()
	sche.Add(Event{
		Id:       id,
		Start:    n + 2*time.Second,
		Interval: time.Second,
		Iterate:  1,
		Task:     &_task{t, &wg},
	})
	id = ai.Id()
	sche.Add(Event{
		Id:       id,
		Start:    n + 2*time.Second,
		Interval: time.Second,
		Iterate:  1,
		Task:     &_task{t, &wg},
	})
	wg.Wait()
}

func TestClose(t *testing.T) {
	sche := New(time.Microsecond * 10)
	sche.ErrorHandler = func(evt Event, err error) {
		t.Errorf("%d, %s", evt.Id, err)
	}
	n := time.Duration(time.Now().UnixNano())

	var wg sync.WaitGroup
	wg.Add(2)
	id := ai.Id()
	sche.Add(Event{
		Id:       id,
		Start:    n + time.Second,
		Interval: time.Second,
		Iterate:  0,
		Task:     &_task{t, &wg},
	})
	id = ai.Id()
	sche.Add(Event{
		Id:       id,
		Start:    n + time.Second,
		Interval: time.Second,
		Iterate:  0,
		Task:     &_task{t, &wg},
	})
	go sche.Serve()
	sche.Close()
	if sche.Count() != 2 {
		t.Errorf("2 Tasks expected, got %d", sche.Count())
	}
}
