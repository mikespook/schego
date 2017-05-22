package schego

import (
	"context"
	"sync"
	"time"

	"github.com/mikespook/golib/idgen"
)

type ExecFunc func(ctx context.Context) error

type event struct {
	sync.RWMutex
	Id        string
	TaskIdGen idgen.IdGen
	Start     time.Time
	Interval  time.Duration
	Iterate   int
	ExecFunc  ExecFunc
}

func (evt event) TaskId() interface{} {
	if evt.TaskIdGen != nil {
		return evt.TaskIdGen.Id()
	}
	return nil
}
