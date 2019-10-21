package driver_test

import (
	"testing"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

var _ gobot.Driver = (*Adc7768Driver)(nil)

func TestAdc7768Driver(t *testing.T) {
	d := NewAdc7768Driver(NewAdc7768Adaptor())

	gobottest.Assert(t, d.Name(), "Adc7768")
	gobottest.Assert(t, d.Connection().Name(), "Adc7768")

	ret := d.Command(Hello)(nil)
	gobottest.Assert(t, ret.(string), "hello from Adc7768!")

	gobottest.Assert(t, d.Ping(), "pong")

	gobottest.Assert(t, len(d.Start()), 0)

	time.Sleep(d.interval)

	sem := make(chan bool, 0)

	d.On(d.Event(Hello), func(data interface{}) {
		sem <- true
	})

	select {
	case <-sem:
	case <-time.After(600 * time.Millisecond):
		t.Errorf("Hello Event was not published")
	}

	gobottest.Assert(t, len(d.Halt()), 0)

	d.On(d.Event(Hello), func(data interface{}) {
		sem <- true
	})

	select {
	case <-sem:
		t.Errorf("Hello Event should not publish after Halt")
	case <-time.After(600 * time.Millisecond):
	}
}
