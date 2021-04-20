package pears_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/peake100/pears-go/pkg/pears"
	"io"
	"math/rand"
	"time"
)

// randNum will be used to do an operation successfully.
var randNum = rand.NewSource(0)

// Catch a panic and return a PanicError from an error value.
func ExampleCatchPanic_panicFromError() {
	// We can use CatchPanic to catch ay panics that occur in an operation
	err := pears.CatchPanic(func() (innerErr error) {
		// We are going to throw an io.EOF.
		panic(io.EOF)
	})

	// Our error will report that it is from a recovered panic.
	fmt.Println("Error:", err)

	// We can test whether this error is a thee result of a panic by using errors.As.
	panicErr := pears.PanicError{}
	if errors.As(err, &panicErr) {
		fmt.Println("error is recovered panic")
		// do something if this was a panic
	}

	// PanicError implements xerrors.Wrapper, so we can use errors.Is and errors.As
	// to get at any inner errors.
	if errors.Is(err, io.EOF) {
		fmt.Println("error is io.EOF")
	}

	// Output:
	// Error: panic recovered: EOF
	// error is recovered panic
	// error is io.EOF
}

// Catch a panic and return a PanicError from a non-error value.
func ExampleCatchPanic_panicFromInt() {
	// We can use CatchPanic to catch ay panics that occur in an operation, even if
	// the panic value is not an error.
	err := pears.CatchPanic(func() (innerErr error) {
		// We are going to throw an io.EOF.
		panic(2)
	})

	fmt.Println("Error:", err)

	// Output:
	// Error: panic recovered: 2
}

// Error passthrough.
func ExampleCatchPanic_errorReturn() {

	// We can use CatchPanic to catch ay panics that occur in an operation
	err := pears.CatchPanic(func() (innerErr error) {
		// We are going to return a normal error.
		return io.EOF
	})

	// We do not get a PanicError this time.
	fmt.Println("Error:", err)

	// Output:
	// Error: EOF
}

// A successful operation wrapped in CatchPanic.
func ExampleCatchPanic_success() {
	// We can use CatchPanic to catch ay panics that occur in an operation
	var result int64
	err := pears.CatchPanic(func() (innerErr error) {
		// We are going to return a normal error.
		result = randNum.Int63()
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("RESULT:", result)

	// Output:
	// RESULT: 8717895732742165505
}

// Have the first error cause all other operations to abort.
func ExampleNewRoutineManager_abortOnError() {
	manager := pears.NewRoutineManager(
		context.Background(), // this context will be used as the parent to all
		// operation contexts

		true, //abortOnError - this will cause any operation
		// error to cancel all other operations

		pears.BatchMatchFirst, // the returned BatchErrors will unwrap to the first
		// error returned fom an operation
	)

	for i := 0; i < 10; i++ {
		// Each routine will be identified as 'worker [workerNum]'. We do not need to
		// use the 'go' keyword here. op will be launched as a routine, but some internal
		// internal bookkeeping needs to occur before the op can be launched.
		workerNum := i
		manager.LaunchRoutine(fmt.Sprint("worker", workerNum), func(ctx context.Context) error {
			// We'll use a timer to stand in for some long-running worker.
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-ctx.Done():
				fmt.Printf("operation %v received abort request\n", workerNum)
				return ctx.Err()
			case <-timer.C:
				fmt.Printf("operation %v completed successfully\n", workerNum)
				return nil
			}
		})
	}

	// Lastly we'll launch a routine that returns an error, which will cancel the
	// contexts of every op launched above.
	manager.LaunchRoutine("faulty operation", func(ctx context.Context) error {
		// This faulty operation will return an io.EOF
		return io.EOF
	})

	// Now we join the manager, which blocks until all routines launched above return.
	// If any operations returned an error, we will get one here.
	err := manager.Join()

	// report our error.
	fmt.Println("\nERROR:", err)

	// errors.Is() and errors.As() can inspect what caused our operations to fail.
	// Because pears.BatchMatchFirst is our error-matching mode, only the FIRST
	// encountered error will pass errors.Is() or errors.As().
	//
	// For us that should be io.EOF.
	if errors.Is(err, io.EOF) {
		fmt.Println("error is io.EOF")
	}

	// Even though the other operations returned context.Canceled, we will NOT
	// pass the following check since it was not the FIRST error returned. This is nice
	// for checking against an error that started a cascade.
	//
	// If our match mode had been set to pears.BatchMatchAny, this check would also
	// pass
	if errors.Is(err, context.Canceled) {
		fmt.Println("error is context.Cancelled")
	}

	// We can extract a pears.OpError to get more information about the first error.
	opErr := pears.OpError{}
	if !errors.As(err, &opErr) {
		panic("expected opErr")
	}

	fmt.Println("batch failure caused by operation:", opErr.OpName)

	// We can also extract a BatchErrors to inspect all if our errors more closely:
	batchErr := pears.BatchErrors{}
	if !errors.As(err, &batchErr) {
		panic("expected BatchErrors")
	}

	// Let's inspect ALL of the errors we got back. We'll see that the context
	// cancellation errors were returned, but because of our Batch error matching mode,
	// are being kept from surfacing through errors.Is() and errors.As().
	fmt.Println("\nALL ERRORS:")
	for _, thisErr := range batchErr.Errs {
		fmt.Println(thisErr)
	}

	// Unordered Output:
	//
	// operation 9 received abort request
	// operation 8 received abort request
	// operation 1 received abort request
	// operation 3 received abort request
	// operation 6 received abort request
	// operation 4 received abort request
	// operation 0 received abort request
	// operation 7 received abort request
	// operation 5 received abort request
	// operation 2 received abort request
	//
	// ERROR: 11 errors returned. first: error during 'faulty operation': EOF
	// error is io.EOF
	// batch failure caused by operation: faulty operation
	//
	// ALL ERRORS:
	// error during 'faulty operation': EOF
	// error during 'worker1': context canceled
	// error during 'worker5': context canceled
	// error during 'worker7': context canceled
	// error during 'worker0': context canceled
	// error during 'worker4': context canceled
	// error during 'worker6': context canceled
	// error during 'worker8': context canceled
	// error during 'worker3': context canceled
	// error during 'worker2': context canceled
	// error during 'worker9': context canceled
}
