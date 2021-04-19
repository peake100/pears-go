package pears_test

import (
	"context"
	"errors"
	"fmt"
	pears "github.com/peake100/pears-go/pkg"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

func TestRoutineManager_NoErrs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Each of our ops will increment this counter.
	counter := 0
	counterLock := new(sync.Mutex)

	manager := pears.NewRoutineManager(ctx, true, pears.BatchMatchFirst)

	for i := 0; i < 10; i++ {
		opIndex := i
		manager.LaunchRoutine(fmt.Sprint("counter", opIndex), func(ctx context.Context) error {
			counterLock.Lock()
			defer counterLock.Unlock()
			counter++

			return nil
		})
	}

	err := manager.Join()
	assert.NoError(t, err, "no errors running routines")
	assert.Equal(t, counter, 10, "counter incremented 10 times")
}

func TestRoutineManager_AllErrs(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager := pears.NewRoutineManager(ctx, true, pears.BatchMatchFirst)

	for i := 0; i < 10; i++ {
		opIndex := i
		manager.LaunchRoutine(fmt.Sprint("op", opIndex), func(ctx context.Context) error {
			return io.EOF
		})
	}

	err := manager.Join()
	if !assert.Error(err, "error running routines") {
		t.FailNow()
	}

	batchErrs := pears.BatchErrors{}
	if !assert.ErrorAs(err, &batchErrs) {
		t.FailNow()
	}

	assert.Len(batchErrs.Errs, 10, "all routines returned errors.")

	for _, err = range batchErrs.Errs {
		opErr := pears.OpError{}
		assert.ErrorAs(err, &opErr, "error is OpErr type")
		assert.ErrorIs(err, io.EOF, "error unwraps to io.EOF")
	}
}

func TestRoutineManager_AbortOnError(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager := pears.NewRoutineManager(ctx, true, pears.BatchMatchFirst)

	// Launch 10 routines that return an error when the context is cancelled, but block
	// before that.
	for i := 0; i < 10; i++ {
		opIndex := i
		manager.LaunchRoutine(fmt.Sprint("op", opIndex), func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
	}

	manager.LaunchRoutine(fmt.Sprint("failOp"), func(ctx context.Context) error {
		return io.EOF
	})

	err := manager.Join()

	batchErrs := pears.BatchErrors{}
	if !assert.ErrorAs(err, &batchErrs) {
		t.FailNow()
	}

	if !assert.Len(batchErrs.Errs, 11, "one routine returned errors.") {
		t.FailNow()
	}

	// Our batch error should only match on the first error (which causes an abort):
	//the io.EoF.
	assert.ErrorIs(batchErrs, io.EOF, "io.EOF is first error")
	assert.False(errors.Is(batchErrs, context.Canceled), "batch error is not Cancelled error")
	assert.False(errors.Is(batchErrs, context.DeadlineExceeded), "batch error is not Cancelled error")

	assert.ErrorIs(batchErrs.Errs[0], io.EOF, "first error is an io.EOF error")
	assert.ErrorIs(batchErrs.Errs[1], context.Canceled, "second error is a cancellation error")

	causingErr := pears.OpError{}
	if !assert.ErrorAs(batchErrs, &causingErr) {
		t.FailNow()
	}

	assert.Equal(causingErr.OpName, "failOp", "first error is OpError from 'failOp' routine")
}

func TestRoutineManager_DoNotAbortOnError(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Each of our ops will increment this counter.
	counter := 0
	counterLock := new(sync.Mutex)

	badOpReturned := make(chan struct{})

	manager := pears.NewRoutineManager(ctx, false, pears.BatchMatchFirst)

	// Launch a routine that returns an error.
	manager.LaunchRoutine(fmt.Sprint("failOp"), func(ctx context.Context) error {
		// Signal to other operations on the way out that we have returned our error
		defer close(badOpReturned)
		return io.EOF
	})

	// Launch 10 routines that return an error when the context is cancelled, but block
	// before that.
	for i := 0; i < 10; i++ {
		opIndex := i
		manager.LaunchRoutine(fmt.Sprint("op", opIndex), func(ctx context.Context) error {
			// Wait to heer our error routine has returned.
			<-badOpReturned

			// Give it 50ms for the error collector to get the error. We want to make
			// sure the context is NOT cancelled in that time.
			timer := time.NewTimer(50 * time.Millisecond)

			select {
			case <-ctx.Done():
				t.Error("context was cancelled before operations exited")
				return ctx.Err()
			case <-timer.C:
			}

			// Increment the counter and return without error.
			counterLock.Lock()
			defer counterLock.Unlock()
			counter++

			return nil
		})
	}

	err := manager.Join()
	if !assert.Error(err, "error running routines") {
		t.FailNow()
	}

	assert.Equal(10, counter, "counter incremented by all workers")

	batchErrs := pears.BatchErrors{}
	if !assert.ErrorAs(err, &batchErrs) {
		t.FailNow()
	}

	if !assert.Len(batchErrs.Errs, 1, "all routines returned errors.") {
		t.FailNow()
	}

	// Our batch error should only match on the first error (which causes an abort):
	//the io.EoF.
	assert.ErrorIs(batchErrs, io.EOF, "io.EOF is first error")
	assert.ErrorIs(batchErrs.Errs[0], io.EOF, "first error is an io.EOF error")

	causingErr := pears.OpError{}
	if !assert.ErrorAs(batchErrs, &causingErr) {
		t.FailNow()
	}

	assert.Equal(causingErr.OpName, "failOp", "first error is OpError from 'failOp' routine")
}

func TestRoutineManager_Join_PanicOnSecondCall(t *testing.T) {
	manager := pears.NewRoutineManager(context.Background(), true, pears.BatchMatchFirst)
	manager.Join()

	assert.Panics(t, func() {
		manager.Join()
	}, "panic on seconds call to Join()")
}

func TestRoutineManager_LaunchRoutine_PanicsOnCallAfterJoin(t *testing.T) {
	manager := pears.NewRoutineManager(context.Background(), true, pears.BatchMatchFirst)
	manager.Join()

	assert.Panics(t, func() {
		manager.LaunchRoutine("panics", func(ctx context.Context) error {
			return nil
		})
	}, "panic on seconds call to Join()")
}
