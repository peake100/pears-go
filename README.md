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
	// We can use CatchPanic to catch ay panics that occur in an operation:
	err := pears.CatchPanic(func() (innerErr error) {
		// We are going to throw an io.EOF.
		panic(io.EOF)
	})

	// Our error will report that it is from a recovered panic.
	fmt.Println("Error:", err)

	// We can test whether this error is a the result of a panic by using errors.As.
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

pears offers a ``Group`` type which takes some inspirations from 
[https://pkg.go.dev/golang.org/x/sync/errgroup](errgroup.Group), with some key 
differences:

- All errors are collected, not just the first. Each is wrapped in an OpError and 
  then collected into a GroupErrors. These types offer a number of ways to inspect
  and resolve errors in concurrent situations.

- Launched operations can be named using GoNamed for more robust error inspection and
  handling.

- A context is required, and is passed to all child functions, allowing for higher
  readability of where a context comes from.

- Group must be created with a constructor function: NewGroup.

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

func main() {
	group := pears.NewGroup(
		context.Background(), // this context will be used as the parent to al
		// operation contexts
	)

	for i := 0; i < 10; i++ {
		// Each routine will be identified as 'worker [workerNum]'. We do not need to
		// use the 'go' keyword here. op will be launched as a routine, but some
		// internal bookkeeping needs to occur before the op can be launched.
		workerNum := i
		group.GoNamed(fmt.Sprint("worker", workerNum), func(ctx context.Context) error {
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
	group.GoNamed("faulty operation", func(ctx context.Context) error {
		// This faulty operation will return an io.EOF
		return io.EOF
	})

	// Now we join the group, which blocks until all routines launched above return.
	// If any operations returned an error, we will get one here.
	err := group.Wait()

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

	// We can also extract a GroupErrors to inspect all of our errors more closely:
	groupErrs := pears.GroupErrors{}
	if !errors.As(err, &groupErrs) {
    panic("expected BatchErrors")
  }

	// Let's inspect ALL of the errors we got back. We'll see that the context
	// cancellation errors were returned, but because of our Batch error matching mode,
	// are being kept from surfacing through errors.Is() and errors.As().
	fmt.Println("\nALL ERRORS:")
	for _, thisErr := range groupErrs.Errs {
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
[read the docs](https://pkg.go.dev/github.com/peake100/pears-go?readme=expanded#section-documentation).

For library development guide, 
[read the docs](https://illuscio-dev.github.io/islelib-go/).

### Prerequisites

Golang 1.6+

## Authors

* **Billy Peake** - *Initial work*

## Attributions

<div>Logo made by <a href="https://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a></div>