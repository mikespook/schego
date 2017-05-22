package schego

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mikespook/golib/idgen"
)

const (
	ForEver = -1
)

var (
	ErrTaskNotFound  = errors.New("The task was not found.")
	ErrTaskIsRunning = errors.New("The task is running.")
)

type Scheduler struct {
	sync.RWMutex
	ticks    map[time.Time][]interface{}
	events   map[interface{}]event
	fireChan chan event
	ticker   *time.Ticker

	ErrorFunc func(context.Context, error)
}

func New(d time.Duration) *Scheduler {
	sche := &Scheduler{
		ticks:    make(map[time.Time][]interface{}),
		events:   make(map[interface{}]event),
		fireChan: make(chan event, 64),
		ticker:   time.NewTicker(d),
	}
	return sche
}

func (sche *Scheduler) Serve() {
	go sche.fire()
	for now := range sche.ticker.C {
		go func(now time.Time) {
			sche.Lock()
			defer sche.Unlock()
			for t := range sche.ticks {
				// task executing time less/equal current time
				if t.Before(now) {
					for index := range sche.ticks[t] {
						id := sche.ticks[t][index]
						if evt, ok := sche.events[id]; ok {
							sche.fireChan <- evt
						}
					}
					delete(sche.ticks, t)
				}
			}
		}(now)
	}
}

func (sche *Scheduler) Close() error {
	sche.ticker.Stop()
	close(sche.fireChan)
	return nil
}

func (sche *Scheduler) fire() {
	for evt := range sche.fireChan {
		go func(evt event) {
			if evt.ExecFunc != nil {
				go func(evt event) {
					ctx := context.WithValue(context.Background(), "event", evt)
					err := evt.ExecFunc(ctx)
					if err != nil {
						sche.err(ctx, err)
					}
				}(evt)
			}
			if evt.Iterate != ForEver {
				evt.Lock()
				defer evt.Unlock()
				if evt.Iterate > 0 {
					evt.Iterate--
				}
				if evt.Iterate == 0 {
					sche.Lock()
					delete(sche.events, evt.Id)
					sche.Unlock()
					return
				}
			}
			evt.Start = evt.Start.Add(evt.Interval)
			sche.add(evt)
		}(evt)
	}
}

func (sche *Scheduler) err(ctx context.Context, err error) {
	if sche.ErrorFunc != nil {
		sche.ErrorFunc(ctx, err)
	}
}

func (sche *Scheduler) Add(id string, start time.Time, interval time.Duration, iterate int, f ExecFunc) {
	sche.Lock()
	defer sche.Unlock()
	evt := event{
		Id:        id,
		TaskIdGen: idgen.NewObjectId(),
		Start:     start,
		Interval:  interval,
		Iterate:   iterate,
		ExecFunc:  f,
	}
	sche.add(evt)
}

func (sche *Scheduler) add(evt event) {
	sche.events[evt.Id] = evt
	if sche.ticks[evt.Start] == nil {
		sche.ticks[evt.Start] = make([]interface{}, 0, 8)
	}
	sche.ticks[evt.Start] = append(sche.ticks[evt.Start], evt.Id)
}
