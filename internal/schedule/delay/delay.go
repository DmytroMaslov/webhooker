package delay

import (
	"time"

	"github.com/hmgle/delaytask"
)

type Delay struct {
	Task *delaytask.Task
}

func NewDelay() *Delay {
	return &Delay{
		Task: delaytask.New(),
	}
}

func (d *Delay) AddJobFn(id string, fn func(), delay time.Duration) {
	d.Task.AddJobFn(id, fn, delay)
}

func (d *Delay) Cancel(id string) {
	d.Task.Cancel(id)
}

func (d *Delay) GracefulExit() <-chan bool {
	return d.Task.GracefulExit()
}
