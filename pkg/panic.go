package pears

import (
	"fmt"
	"runtime/debug"
)

// PanicErr is returned by CatchPanic when a panic was recovered.
type PanicErr struct {
	// Recovered contains the original value recovered from recovered().
	Recovered interface{}
	// RecoveredErr is Recovered converted to an error, through a type assertion if
	// Recovered implements error, or fmt.Errorf("%v", Recovered) if it does not.
	RecoveredErr error
	// StackTrace contains the formatted stacktrace of the panic.
	StackTrace string
}

// Error implements builtins.error.
func (err PanicErr) Error() string {
	return fmt.Sprint("panic recovered: ", err.RecoveredErr)
}

// Unwrap implements xerrors.Wrapper for unwraps to RecoveredErr.
func (err PanicErr) Unwrap() error {
	return err.RecoveredErr
}

// CatchPanic runs mayPanic and returns it's result.
//
// If mayPanic panics, the panic is recovered and a PanicErr is returned with the
// recovered value.
func CatchPanic(mayPanic func() (innerErr error)) (err error) {
	// Defer catching a panic.
	defer func() {
		var stacktrace []byte
		var recovered interface{}
		// If there is nothing to recover, return.
		if recovered = recover(); recovered == nil {
			return
		}

		stacktrace = debug.Stack()

		// Check if the recovered value is an error.
		var recoveredErr error
		var ok bool
		if recoveredErr, ok = recovered.(error); !ok {
			// If it is not, convert it to one.
			recoveredErr = fmt.Errorf("%v", recovered)
		}

		// Set the return error to a PanicErr.
		err = PanicErr{
			Recovered:    recovered,
			RecoveredErr: recoveredErr,
			StackTrace:   string(stacktrace),
		}
	}()

	// Run the caller's function.
	err = mayPanic()
	return err
}