package schego

import (
	"time"

	"github.com/mikespook/golib/idgen"
)

type Event struct {
	Id        interface{}
	TaskIdGen idgen.IdGen
	Start     time.Duration
	Interval  time.Duration
	Iterate   int
	Task      Task
}

func (evt Event) TaskId() interface{} {
	if evt.TaskIdGen != nil {
		return evt.TaskIdGen.Id()
	}
	return nil
}
