package schego

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrTaskNotFound  = errors.New("The task was not found.")
	ErrTaskIsRunning = errors.New("The task is running.")
)

type Scheduler struct {
	sync.RWMutex
	ticks       map[time.Duration][]interface{}
	events      map[interface{}]Event
	tick        time.Duration
	fireChan    chan Event
	HandleError func(Event, error)
}

func New(tick time.Duration) *Scheduler {
	sche := &Scheduler{
		ticks:    make(map[time.Duration][]interface{}),
		events:   make(map[interface{}]Event),
		tick:     tick,
		fireChan: make(chan Event, 64),
	}
	return sche
}

func (sche *Scheduler) Loop() {
	go sche.fire()
	for now := range time.Tick(sche.tick) {
		go func(now time.Time) {
			current := time.Duration(now.UnixNano())
			sche.Lock()
			defer sche.Unlock()
			for t := range sche.ticks {
				// task executing time less/equal current time
				if t <= current {
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

func (sche *Scheduler) fire() {
	for evt := range sche.fireChan {
		go func(evt Event) {
			if evt.Task != nil {
				err := evt.Task.Exec()
				if err != nil {
					sche.err(evt, err)
				}
			}
			if evt.Iterate <= 0 {
				sche.Lock()
				delete(sche.events, evt.Id)
				sche.Unlock()
			} else {
				evt.Start += evt.Interval
				evt.Iterate--
				sche.Add(evt)
			}
		}(evt)
	}
}

func (sche *Scheduler) err(evt Event, err error) {
	if sche.HandleError != nil {
		sche.HandleError(evt, err)
	}
}

func (sche *Scheduler) Add(evt Event) {
	sche.Lock()
	defer sche.Unlock()
	sche.events[evt.Id] = evt
	if sche.ticks[evt.Start] == nil {
		sche.ticks[evt.Start] = make([]interface{}, 0, 8)
	}
	sche.ticks[evt.Start] = append(sche.ticks[evt.Start], evt.Id)
}

func (sche *Scheduler) Remove(id interface{}) {
	sche.Lock()
	defer sche.Unlock()
	delete(sche.events, id)
}

func (sche *Scheduler) Cancel(id interface{}) (err error) {
	sche.Lock()
	defer sche.Unlock()
	if evt, ok := sche.events[id]; ok {
		err = evt.Task.Cancel()
		delete(sche.events, id)
		return
	}
	return ErrTaskNotFound
}

func (sche *Scheduler) Exec(id interface{}) (err error) {
	sche.Lock()
	defer sche.Unlock()
	if evt, ok := sche.events[id]; ok {
		err = evt.Task.Exec()
		delete(sche.events, id)
		return
	}
	return ErrTaskNotFound
}

func (sche *Scheduler) Get(id interface{}) Event {
	sche.RLock()
	defer sche.RUnlock()
	if evt, ok := sche.events[id]; ok {
		return evt
	}
	return Event{}
}

func (sche *Scheduler) Count() int {
	sche.RLock()
	defer sche.RUnlock()
	return len(sche.events)
}

func (sche *Scheduler) TickCount(t time.Duration) int {
	sche.RLock()
	defer sche.RUnlock()
	return len(sche.ticks[t])
}
