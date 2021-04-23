package pears

import (
	"context"
	"sync"
	"sync/atomic"
)

// Group runs a number of concurrent operations and collects their errors.
//
// Group takes some inspirations from: https://pkg.go.dev/golang.org/x/sync/errgroup,
// with some key differences:
//
// - All errors are collected, not just the first. Each is wrapped in an OpError and
//   then collected into a GroupErrors.
//
// - Launched operations can be named using GoNamed for more robust error inspection and
//   handling.
//
// - A context is required, and is passed to all child functions, allowing for higher
//   readability.
//
// - Group must be created with a constructor function: NewGroup.
type Group struct {
	// ctx is the main context we will pass to all Go ops.
	ctx context.Context
	// cancel cancels ctx.
	cancel context.CancelFunc

	// joined will be atomically inspected by Go and Wait to see if Wait has
	// already been called.
	//
	// 1 = Wait has been called.
	joined int32

	// opErrors receives errors encountered by ops run in Go for collection.
	opErrors chan error
	// opsDone will be added to for every call to Go before returning, and
	// waited on before Wait exits.
	opsDone sync.WaitGroup
	// joinCalled will be closed when Wait is called and signals to the error collection
	// routine to start wrapping up.
	joinCalled chan struct{}
	// errorsCollected will be closed by the error collection routine once all errors
	// are collected to signal that Wait can process them.
	errorsCollected chan struct{}

	// SETTINGS --------

	// abortOnErr will cause cancel to be called as soon as a routine op returns an
	// error.
	abortOnErr bool
	// errMode is
	errMode GroupMatchMode

	// RESULTS ----------

	// collectedErrs stores opErrors as they are returned by operations.
	collectedErrs []error
}

// collectErrors will be run in it's own routine and collect the errors from our
// workers.
func (runner *Group) collectErrors() {
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

	// Wait for Wait() to be called.
	<-runner.joinCalled

	// Wait for all operations to be done.
	runner.opsDone.Wait()

	// Close the error channel.
	close(runner.opErrors)

	// Wait for collection to be done.
	<-collectionDone
}

// Go launches op in it's own routine and sends any returned errors to be
// collected.
//
// Go will panic if called after Wait.
func (runner *Group) Go(op func(ctx context.Context) error) {
	runner.GoNamed("[ROUTINE]", op)
}

// GoNamed can be used to give your routine a name.
func (runner *Group) GoNamed(name string, op func(ctx context.Context) error) {
	if !atomic.CompareAndSwapInt32(&runner.joined, 0, 0) {
		panic("Group.Go called after Group.Wait")
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

// Wait waits until all operations launched by Go complete. If any errors
// are returned by operations, they will be returned as OpError values in a
// GroupErrors container.
//
// Wait will panic if called multiple times. Group cannot be reused.
func (runner *Group) Wait() error {
	defer runner.cancel()
	if !atomic.CompareAndSwapInt32(&runner.joined, 0, 1) {
		panic("Group.Wait called multiple times")
	}
	close(runner.joinCalled)

	<-runner.errorsCollected
	if len(runner.collectedErrs) == 0 {
		return nil
	}

	return GroupErrors{
		MatchMode: runner.errMode,
		Errs:      runner.collectedErrs,
	}
}

// NewGroup creates a new *Group for running concurrent operations and
// centrally collecting their errors.
//
// ctx is the main context.Context for the batch. It will be used as the parent context
// of the ctx parameter passed to all Group.Go op functions.
//
// The returned group can be configured with opts.
//
// Options and defaults:
//
// - WithAbortOnError: true
//
// - WithErrMode: GroupMatchFirst
func NewGroup(
	ctx context.Context,
	opts ...GroupOption,
) *Group {
	managerCtx, cancel := context.WithCancel(ctx)

	// create the routine group.
	group := &Group{
		ctx:             managerCtx,
		cancel:          cancel,
		joined:          0,
		opErrors:        make(chan error, 1),
		opsDone:         sync.WaitGroup{},
		joinCalled:      make(chan struct{}),
		errorsCollected: make(chan struct{}),
		abortOnErr:      true,
		errMode:         GroupMatchFirst,
		collectedErrs:   make([]error, 0),
	}

	// Apply our options.
	for _, opt := range opts {
		opt(group)
	}

	// Launch the error collection routine.
	go group.collectErrors()

	// Return the group to the caller.
	return group
}

// GroupOption defines an option for Group.
type GroupOption = func(group *Group)

// WithAbortOnError sets whether the Group should abort on the first encountered error.
//
// Default: true.
func WithAbortOnError(abort bool) GroupOption {
	return func(group *Group) {
		group.abortOnErr = abort
	}
}

// WithErrMode returns the error mode that will be passed to the returned GroupErr.
//
// Default: GroupMatchFirst.
func WithErrMode(mode GroupMatchMode) GroupOption {
	return func(group *Group) {
		group.errMode = mode
	}
}
