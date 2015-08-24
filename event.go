package schego

import (
	"time"
)

type Event struct {
	Id       interface{}
	Start    time.Duration
	Interval time.Duration
	Iterate  int
	Task     Task
}
