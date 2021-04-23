package pears_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/peake100/pears-go/pkg/pears"
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

	manager := pears.NewGroup(ctx)

	for i := 0; i < 10; i++ {
		opIndex := i
		manager.GoNamed(fmt.Sprint("counter", opIndex), func(ctx context.Context) error {
			counterLock.Lock()
			defer counterLock.Unlock()
			counter++

			return nil
		})
	}

	err := manager.Wait()
	assert.NoError(t, err, "no errors running routines")
	assert.Equal(t, counter, 10, "counter incremented 10 times")
}

func TestRoutineManager_AllErrs(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager := pears.NewGroup(ctx)

	for i := 0; i < 10; i++ {
		manager.Go(func(ctx context.Context) error {
			return io.EOF
		})
	}

	err := manager.Wait()
	if !assert.Error(err, "error running routines") {
		t.FailNow()
	}

	batchErrs := pears.GroupErrors{}
	if !assert.ErrorAs(err, &batchErrs) {
		t.FailNow()
	}

	assert.Len(batchErrs.Errs, 10, "all routines returned errors.")

	for _, err = range batchErrs.Errs {
		opErr := pears.OpError{}
		if assert.ErrorAs(err, &opErr, "error is OpErr type") {
			assert.Equal("[ROUTINE]", opErr.OpName, "default name expected")
		}
		assert.ErrorIs(err, io.EOF, "error unwraps to io.EOF")
	}
}

func TestRoutineManager_AbortOnError(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager := pears.NewGroup(ctx)

	// Launch 10 routines that return an error when the context is cancelled, but block
	// before that.
	for i := 0; i < 10; i++ {
		opIndex := i
		manager.GoNamed(fmt.Sprint("op", opIndex), func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
	}

	manager.GoNamed(fmt.Sprint("failOp"), func(ctx context.Context) error {
		return io.EOF
	})

	err := manager.Wait()

	batchErrs := pears.GroupErrors{}
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

	manager := pears.NewGroup(ctx, pears.WithAbortOnError(false))

	// Launch a routine that returns an error.
	manager.GoNamed(fmt.Sprint("failOp"), func(ctx context.Context) error {
		// Signal to other operations on the way out that we have returned our error
		defer close(badOpReturned)
		return io.EOF
	})

	// Launch 10 routines that return an error when the context is cancelled, but block
	// before that.
	for i := 0; i < 10; i++ {
		opIndex := i
		manager.GoNamed(fmt.Sprint("op", opIndex), func(ctx context.Context) error {
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

	err := manager.Wait()
	if !assert.Error(err, "error running routines") {
		t.FailNow()
	}

	assert.Equal(10, counter, "counter incremented by all workers")

	batchErrs := pears.GroupErrors{}
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

func TestRoutineManager_Wait_PanicOnSecondCall(t *testing.T) {
	manager := pears.NewGroup(context.Background())
	manager.Wait()

	assert.Panics(t, func() {
		manager.Wait()
	}, "panic on seconds call to Wait()")
}

func TestRoutineManager_LaunchRoutine_PanicsOnCallAfterWait(t *testing.T) {
	manager := pears.NewGroup(context.Background())
	manager.Wait()

	assert.Panics(t, func() {
		manager.GoNamed("panics", func(ctx context.Context) error {
			return nil
		})
	}, "panic on seconds call to Wait()")
}

func TestRoutineManager_GroupMatchAll(t *testing.T) {
	assert := assert.New(t)

	manager := pears.NewGroup(context.Background(), pears.WithErrMode(pears.GroupMatchAny))

	manager.GoNamed("returns io.EOF", func(ctx context.Context) error {
		return io.EOF
	})

	manager.GoNamed("returns io.ErrClosedPipe", func(ctx context.Context) error {
		return io.ErrClosedPipe
	})

	err := manager.Wait()
	if !assert.Error(err, "Wait returns error") {
		t.FailNow()
	}

	assert.ErrorIs(err, io.EOF)
	assert.ErrorIs(err, io.ErrClosedPipe)
}
