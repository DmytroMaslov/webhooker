package delay

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Delay_Func(t *testing.T) {

	delay := NewDelay()

	a := 1

	go delay.AddJobFn("test", func() {
		a++
	}, time.Millisecond)

	time.Sleep(5 * time.Millisecond)

	assert.Equal(t, 2, a)
}

func Test_Delay_Cancel(t *testing.T) {
	delay := NewDelay()

	a := 1

	go delay.AddJobFn("test", func() {
		a++
	}, time.Millisecond)

	delay.Cancel("test")

	assert.Equal(t, 1, a)
}

func Test_Delay_ShotDown(t *testing.T) {
	delay := NewDelay()

	a := 1

	go delay.AddJobFn("test", func() {
		a++
	}, time.Second)
	time.Sleep(100 * time.Millisecond)

	<-delay.GracefulExit()

	assert.Equal(t, 2, a)
}
