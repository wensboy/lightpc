package async

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFutureGetReturnsResultAndError(t *testing.T) {
	value, err := New[int](nil, func(context.Context) (int, error) {
		return 7, nil
	}).Get()
	if err != nil || value != 7 {
		t.Fatalf("value = %v, err = %v", value, err)
	}

	sentinel := errors.New("boom")
	if _, err := New[int](nil, func(context.Context) (int, error) {
		return 0, sentinel
	}).Get(); !errors.Is(err, sentinel) {
		t.Fatalf("err = %v", err)
	}
}

func TestFuturePanicBecomesError(t *testing.T) {
	_, err := New[int](nil, func(context.Context) (int, error) {
		panic("boom")
	}).Get()
	if err == nil || !strings.Contains(err.Error(), "panic") {
		t.Fatalf("err = %v", err)
	}
}

func TestFutureCancelCancelsContext(t *testing.T) {
	started := make(chan struct{})
	future := New[int](nil, func(ctx context.Context) (int, error) {
		close(started)
		<-ctx.Done()
		return 0, ctx.Err()
	})
	<-started

	future.Cancel()

	if _, err := future.Get(); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestFutureGetContextTimeout(t *testing.T) {
	release := make(chan struct{})
	future := New[int](nil, func(context.Context) (int, error) {
		<-release
		return 42, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := future.GetContext(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}

	close(release)
	if value, err := future.Get(); err != nil || value != 42 {
		t.Fatalf("value = %v, err = %v", value, err)
	}
}

func TestFutureTryGet(t *testing.T) {
	release := make(chan struct{})
	future := New[int](nil, func(context.Context) (int, error) {
		<-release
		return 9, nil
	})

	if _, _, ok := future.TryGet(); ok {
		t.Fatal("TryGet must report not done before completion")
	}

	close(release)
	<-future.Done()

	if value, err, ok := future.TryGet(); !ok || err != nil || value != 9 {
		t.Fatalf("value = %v, err = %v, ok = %v", value, err, ok)
	}
}

func TestFutureGetIsRepeatable(t *testing.T) {
	future := New[string](nil, func(context.Context) (string, error) {
		return "done", nil
	})

	for index := 0; index < 3; index++ {
		if value, err := future.Get(); err != nil || value != "done" {
			t.Fatalf("value = %v, err = %v", value, err)
		}
	}
}

func TestScopeWaitCollectsTasks(t *testing.T) {
	var mu sync.Mutex
	count := 0
	scope := NewScope(nil)
	for index := 0; index < 8; index++ {
		scope.Go(func(context.Context) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})
	}

	if err := scope.Wait(); err != nil {
		t.Fatal(err)
	}
	if count != 8 {
		t.Fatalf("count = %d", count)
	}
}

func TestScopeFirstErrorCancelsSiblings(t *testing.T) {
	sentinel := errors.New("boom")
	started := make(chan struct{})
	scope := NewScope(nil)

	scope.Go(func(context.Context) error {
		close(started)
		return sentinel
	})
	scope.Go(func(ctx context.Context) error {
		<-started
		<-ctx.Done()
		return ctx.Err()
	})

	if err := scope.Wait(); !errors.Is(err, sentinel) {
		t.Fatalf("err = %v", err)
	}
}

func TestScopePanicBecomesError(t *testing.T) {
	scope := NewScope(nil)
	scope.Go(func(context.Context) error {
		panic("boom")
	})

	err := scope.Wait()
	if err == nil || !strings.Contains(err.Error(), "scope panic") {
		t.Fatalf("err = %v", err)
	}
}

func TestScopeContextCancelledAfterWait(t *testing.T) {
	scope := NewScope(nil)
	ctx := scope.Context()

	if err := scope.Wait(); err != nil {
		t.Fatal(err)
	}
	if ctx.Err() == nil {
		t.Fatal("the scope context must be cancelled after Wait")
	}
}

func TestAllKeepsOrder(t *testing.T) {
	values, err := All(
		context.Background(),
		func(context.Context) (int, error) { return 1, nil },
		func(context.Context) (int, error) { return 2, nil },
		func(context.Context) (int, error) { return 3, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(values, []int{1, 2, 3}) {
		t.Fatalf("values = %v", values)
	}
}

func TestAllReturnsFirstError(t *testing.T) {
	sentinel := errors.New("boom")
	values, err := All(
		context.Background(),
		func(context.Context) (int, error) { return 1, nil },
		func(context.Context) (int, error) { return 0, sentinel },
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v", err)
	}
	if values != nil {
		t.Fatalf("values = %v", values)
	}
}

func TestRaceReturnsFirstResult(t *testing.T) {
	value, err := Race(
		context.Background(),
		func(context.Context) (int, error) { return 1, nil },
		func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		},
	)
	if err != nil || value != 1 {
		t.Fatalf("value = %v, err = %v", value, err)
	}
}

func TestRaceWithoutTasks(t *testing.T) {
	value, err := Race[int](context.Background())
	if err != nil || value != 0 {
		t.Fatalf("value = %v, err = %v", value, err)
	}
}

func TestRacePanicBecomesError(t *testing.T) {
	_, err := Race(
		context.Background(),
		func(context.Context) (int, error) {
			panic("boom")
		},
	)
	if err == nil || !strings.Contains(err.Error(), "race panic") {
		t.Fatalf("err = %v", err)
	}
}
