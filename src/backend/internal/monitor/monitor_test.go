package monitor

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMonitorTicks(t *testing.T) {
	ticks := make(chan any, 8)
	monitor := New(time.Millisecond, func() (any, error) {
		return "tick", nil
	}, func(value any) {
		ticks <- value
	}, nil)

	monitor.Start()
	defer monitor.Stop()

	for index := 0; index < 3; index++ {
		select {
		case value := <-ticks:
			if value != "tick" {
				t.Fatalf("value = %v", value)
			}
		case <-time.After(time.Second):
			t.Fatal("the monitor must tick")
		}
	}
}

func TestMonitorReportsHandlerError(t *testing.T) {
	sentinel := errors.New("boom")
	failures := make(chan error, 8)
	monitor := New(time.Millisecond, func() (any, error) {
		return nil, sentinel
	}, nil, func(err error) {
		failures <- err
	})

	monitor.Start()
	defer monitor.Stop()

	select {
	case err := <-failures:
		if !errors.Is(err, sentinel) {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("the failure callback must receive the handler error")
	}
}

func TestMonitorReportsRecoveredPanic(t *testing.T) {
	failures := make(chan error, 8)
	monitor := New(time.Millisecond, func() (any, error) {
		panic("boom")
	}, nil, func(err error) {
		failures <- err
	})

	monitor.Start()
	defer monitor.Stop()

	select {
	case err := <-failures:
		if !strings.Contains(err.Error(), "panicked") {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("a panic must reach the failure callback")
	}
}

func TestMonitorPicksExactlyOneCallback(t *testing.T) {
	var index atomic.Int32
	var successes atomic.Int32
	var failures atomic.Int32

	monitor := New(time.Millisecond, func() (any, error) {
		if index.Add(1)%2 == 0 {
			return nil, errors.New("boom")
		}
		return "ok", nil
	}, func(any) {
		successes.Add(1)
	}, func(error) {
		failures.Add(1)
	})

	monitor.Start()
	time.Sleep(20 * time.Millisecond)
	monitor.Stop()

	if successes.Load() == 0 || failures.Load() == 0 {
		t.Fatalf("successes = %d, failures = %d", successes.Load(), failures.Load())
	}
	if successes.Load()+failures.Load() != index.Load() {
		t.Fatalf("successes = %d, failures = %d, ticks = %d", successes.Load(), failures.Load(), index.Load())
	}
}

func TestMonitorToleratesMissingCallbacks(t *testing.T) {
	monitor := New(time.Millisecond, func() (any, error) {
		return nil, errors.New("boom")
	}, nil, nil)

	monitor.Start()
	time.Sleep(10 * time.Millisecond)
	monitor.Stop()
}

func TestMonitorStopHaltsTicks(t *testing.T) {
	var count atomic.Int32
	monitor := New(time.Millisecond, func() (any, error) {
		count.Add(1)
		return nil, nil
	}, nil, nil)

	monitor.Start()
	time.Sleep(20 * time.Millisecond)
	monitor.Stop()

	stopped := count.Load()
	time.Sleep(20 * time.Millisecond)
	if count.Load() != stopped {
		t.Fatal("ticks must stop after Stop")
	}
}

func TestMonitorStopIsIdempotent(t *testing.T) {
	monitor := New(time.Millisecond, func() (any, error) {
		return nil, nil
	}, nil, nil)

	monitor.Stop()
	monitor.Start()
	monitor.Stop()
	monitor.Stop()
}

func TestMonitorStartIsIdempotent(t *testing.T) {
	var count atomic.Int32
	monitor := New(time.Millisecond, func() (any, error) {
		count.Add(1)
		return nil, nil
	}, nil, nil)

	monitor.Start()
	monitor.Start()
	time.Sleep(20 * time.Millisecond)

	// A single Stop must silence the monitor; a second loop would keep ticking.
	monitor.Stop()
	stopped := count.Load()
	time.Sleep(20 * time.Millisecond)
	if count.Load() != stopped {
		t.Fatal("a duplicated Start must not spawn a second loop")
	}
}

func TestNewFallsBackToDefaultInterval(t *testing.T) {
	monitor := New(0, func() (any, error) {
		return nil, nil
	}, nil, nil)

	if monitor.interval != defaultInterval {
		t.Fatalf("interval = %v", monitor.interval)
	}
}

func TestMonitorConcurrentStartStop(t *testing.T) {
	monitor := New(time.Millisecond, func() (any, error) {
		return nil, nil
	}, nil, nil)

	var waitGroup sync.WaitGroup
	for index := 0; index < 4; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			monitor.Start()
			monitor.Stop()
		}()
	}
	waitGroup.Wait()

	monitor.Stop()
}
