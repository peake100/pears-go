<h1 align="center">Pears</h1>
<p align="center">
    <img height=200 align="center" src="https://raw.githubusercontent.com/peake100/pears-go/main/zdocs/source/_static/logo.svg"/>
</p>
<p align="center">Harvest Go Errors with Ease</p>
<p align="center">
    <a href="https://dev.azure.com/peake100/Peake100/_build?definitionId=10"><img src="https://dev.azure.com/peake100/Peake100/_apis/build/status/pears-go?repoName=peake100%2Fpears-go&branchName=dev" alt="click to see build pipeline"></a>
    <a href="https://dev.azure.com/peake100/Peake100/_build?definitionId=10"><img src="https://img.shields.io/azure-devops/tests/peake100/Peake100/10/dev?compact_message" alt="click to see build pipeline"></a>
    <a href="https://dev.azure.com/peake100/Peake100/_build?definitionId=10"><img src="https://img.shields.io/azure-devops/coverage/peake100/Peake100/10/dev?compact_message" alt="click to see build pipeline"></a>
</p>
<p align="center">
    <a href="https://goreportcard.com/report/github.com/illuscio-dev/islelib-go"><img src="https://goreportcard.com/badge/github.com/illuscio-dev/islelib-go" alt="click to see report card"></a>
    <a href="https://codeclimate.com/github/peake100/pears-go/maintainability"><img src="https://api.codeclimate.com/v1/badges/eb73ef0f82b6bd72d7a2/maintainability" /></a>
</p>
<p align="center">
    <a href="https://github.com/peake100/pears-go"><img src="https://img.shields.io/github/go-mod/go-version/peake100/pears-go" alt="Repo"></a>
    <a href="https://pkg.go.dev/github.com/peake100/pears-go?readme=expanded#section-documentation"><img src="https://pkg.go.dev/badge/github.com/peake100/pears-go?readme=expanded#section-documentation.svg" alt="Go Reference"></a>
</p>

Introduction
------------

Pears helps reduce the boilerplate and ensure correctness for common error-handling 
scenarios:

- Panic recovery

- Abort and error collection from concurrent workers.

Demo
----

**Catch a Panic**

```go
package main

import (
	"errors"
	"fmt"
	"github.com/peake100/pears-go/pkg/pears"
	"io"
)

func main() {
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
```

**Gather Errors From Multiple Workers**

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/peake100/pears-go/pkg/pears"
	"io"
	"time"
)

// Have the first error cause all other operations to abort.
func main() {
	manager := pears.NewRoutineManager(
		context.Background(), // this context will be used as the parent to all
		// operation contexts

		true, //abortOnError - this will cause any operation
		// error to cancel all other operations

		pears.BatchMatchFirst, // the returned BatchErrors will unwrap to the first
		// error returned fom an operation
	)

	for i := 0; i < 10; i++ {
		// Each routine will be identified as 'worker [i]'. We do not need tpo use the
		// 'go' keyword here. Internally, op will be launched as a routine, but
		// LaunchRoutine has to add to an internal WaitGroup before the op can be
		// launched.
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

	// Lastly we'll launch a routine that returns an error. This will cause the ctx of
	// all thee routines launched above to cancel and those routines to abort.
	manager.LaunchRoutine("faulty operation", func(ctx context.Context) error {
		// This faulty operation will return an io.EOF
		return io.EOF
	})

	// Now we join the manager, which blocks until all routines launched above return.
	// We will get back an error if any operations returned an error.
	err := manager.Join()

	// report our error.
	fmt.Println("\nERROR:", err)

	// We can use errors.Is() and errors.As() to inspect what caused our operations
	// to fail. Because pears.BatchMatchFirst was used as our error-matching mode,
	// only the FIRST encountered error will pass errors.Is() or errors.As().
	//
	// For us that should be io.EOF.
	if errors.Is(err, io.EOF) {
		fmt.Println("error is io.EOF")
	}

	// Even though the other operations returned a context.Canceled, we will NOT
	// pass the following check since it was not the FIRST error returned. This makes
	// some sense since these cancellation errors did not really cause the error, they
	// resulted from it.
	if errors.Is(err, context.Canceled) {
		fmt.Println("error is context.Cancelled")
	}

	// We can extract an OpError to get more information about the first error.
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
	// ERROR: 11 errors returned, including: error during faulty operation: EOF
	// error is io.EOF
	// batch failure caused by operation: faulty operation
	//
	// ALL ERRORS:
	// error during faulty operation: EOF
	// error during worker9: context canceled
	// error during worker8: context canceled
	// error during worker1: context canceled
	// error during worker3: context canceled
	// error during worker6: context canceled
	// error during worker4: context canceled
	// error during worker0: context canceled
	// error during worker7: context canceled
	// error during worker5: context canceled
	// error during worker2: context canceled
}
```

Goals
-----

- Expose simple APIs for dealing with common error-handling situations.

- Support error inspection through errors.Is and errors.As,

Non-Goals
---------

- Creating complex error frameworks. Pears does not want to re-invent the wheel and 
  seeks only to reduce the boilerplate of leveraging Go's built-in error system.
  
- Solving niche problems. This package seeks to help only the most-broad error cases.
  Features like HTTP or gRPC error-code and serialization systems are beyond the scope 
  of this package.

## Getting Started
For API documentation:
[read the docs](https://illuscio-dev.github.io/islelib-go/).

For library development guide, 
[read the docs](https://illuscio-dev.github.io/islelib-go/).

### Prerequisites

Golang 1.6+, Python 3.6+

## Authors

* **Billy Peake** - *Initial work*

## Attributions

<div>Logo made by <a href="https://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a></div>