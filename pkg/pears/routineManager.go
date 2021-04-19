package pears

import (
	"context"
	"sync"
	"sync/atomic"
)

// RoutineManager runs a number of concurrent operations and collects their errors.
type RoutineManager struct {
	// SYNCHRONIZATION ---------

	// ctx is the main context we will pass to all LaunchRoutine ops.
	ctx context.Context
	// cancel cancels ctx.
	cancel context.CancelFunc

	// joined will be atomically inspected by LaunchRoutine and Join to see if Join has
	// already been called.
	//
	// 1 = Join has been called.
	joined *int32

	// opErrors receives errors encountered by ops run in LaunchRoutine for collection.
	opErrors chan error
	// opsDone will be added to for every call to LaunchRoutine before returning, and
	// waited on before Join exits.
	opsDone *sync.WaitGroup
	// joinCalled will be closed when Join is called and signals to the error collection
	// routine to start wrapping up.
	joinCalled chan struct{}
	// errorsCollected will be closed by the error collection routine once all errors
	// are collected to signal that Join can process them.
	errorsCollected chan struct{}

	// SETTINGS --------

	// abortOnErr will cause cancel to be called as soon as a routine op returns an
	// error.
	abortOnErr bool
	// errMode is
	errMode BatchMatchMode

	// RESULTS ----------

	// collectedErrs stores opErrors as they are returned by operations.
	collectedErrs []error
}

func (runner *RoutineManager) collectErrors() {
	// Signal we are done collecting errors on exit.
	defer close(runner.errorsCollected)

	// Run a routine collecting errors from operations as they complete.
	collectionDone := make(chan struct{})
	go func() {
		defer close(collectionDone)
		for err := range runner.opErrors {
			runner.collectedErrs = append(runner.collectedErrs, err)
			if runner.abortOnErr {
				runner.cancel()
			}
		}
	}()

	// Wait for Join() to be called.
	<-runner.joinCalled

	// Wait for all operations to be done.
	runner.opsDone.Wait()

	// Close the error channel.
	close(runner.opErrors)

	// Wait for collection to be done.
	<-collectionDone
}

// LaunchRoutine launches op in it's own routine and sends any returned errors to be
// collected.
//
// LaunchRoutine will panic if called after Join.
func (runner *RoutineManager) LaunchRoutine(name string, op func(ctx context.Context) error) {
	if !atomic.CompareAndSwapInt32(runner.joined, 0, 0) {
		panic("RoutineManager.LaunchRoutine called after RoutineManager.Join")
	}

	runner.opsDone.Add(1)

	go func() {
		defer runner.opsDone.Done()
		err := op(runner.ctx)
		if err == nil {
			return
		}

		// Wrap this error in a batch error.
		err = OpError{
			OpName: name,
			Err:    err,
		}

		runner.opErrors <- err
	}()
}

// Join waits until all operations launched by LaunchRoutine complete. If any errors
// are returned by operations, they will be returned as OpError values in a
// BatchErrors container.
//
// Join will panic if called multiple times. RoutineManager cannot be reused.
func (runner *RoutineManager) Join() error {
	defer runner.cancel()
	if !atomic.CompareAndSwapInt32(runner.joined, 0, 1) {
		panic("RoutineManager.Join called multiple times")
	}
	close(runner.joinCalled)

	<-runner.errorsCollected
	if len(runner.collectedErrs) == 0 {
		return nil
	}

	return BatchErrors{
		MatchMode: runner.errMode,
		Errs:      runner.collectedErrs,
	}
}

// NewRoutineManager creates a new *RoutineManager for running concurrent operations and
// centrally collecting their errors.
//
// ctx is the main context.Context for the batch. It will be used as the parent context
// of the ctx parameter passed to all RoutineManager.LaunchRoutine op functions.
//
// abortOnError, when true, will cause the cancellation of all op contexts as soon as
// an error is returned by any op.
//
// batchErrorMode is the BatchMatchMode assigned to the BatchErrors value returned by
// Join when one or more operations returned errors.
func NewRoutineManager(
	ctx context.Context,
	abortOnError bool,
	batchErrorMode BatchMatchMode,
) *RoutineManager {
	managerCtx, cancel := context.WithCancel(ctx)
	closed := int32(0)

	// create the routine manager.
	manager := &RoutineManager{
		ctx:             managerCtx,
		cancel:          cancel,
		joined:          &closed,
		opErrors:        make(chan error, 1),
		opsDone:         new(sync.WaitGroup),
		joinCalled:      make(chan struct{}),
		errorsCollected: make(chan struct{}),
		abortOnErr:      abortOnError,
		errMode:         batchErrorMode,
		collectedErrs:   make([]error, 0),
	}

	// Launch the error collection routine.
	go manager.collectErrors()

	// Return the manager to the caller.
	return manager
}
