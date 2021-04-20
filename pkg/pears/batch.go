package pears

import (
	"errors"
	"fmt"
)

// BatchMatchMode determines how BatchErrors should unwrap.
type BatchMatchMode int

const (
	// BatchMatchNone tells BatchErrors to match errors.Is or errors.As to not match
	// on any error. errors.As will still match on a OpError.
	//
	// BatchErrors.Unwrap will return nil in this mode.
	BatchMatchNone BatchMatchMode = iota
	// BatchMatchAny tells BatchErrors to match errors.Is or errors.As on any contained
	// error. BatchErrors.Unwrap will return the first error in this mode.
	BatchMatchAny
	// BatchMatchFirst tells BatchErrors to unwrap to the first returned error.
	BatchMatchFirst
)

// OpError is a single error returned by a batch operation.
type OpError struct {
	// OpName is the name of the operation this error occurred on.
	OpName string
	// Err is the error returned by the operation.
	Err error
}

// Error implements builtins.error.
func (err OpError) Error() string {
	return fmt.Sprintf(
		"error during '%v': %v", err.OpName, err.Err,
	)
}

// Unwrap implements xerrors.Wrapper.
func (err OpError) Unwrap() error {
	return err.Err
}

// BatchErrors is a group of errors
type BatchErrors struct {
	// MatchMode indicates how Is, As, and Unwrap should behave.
	//
	// BatchMatchNone: Is will always return false, As will only return true if the
	// target is a *OpError, and Unwrap will return nil if called directly.
	//
	// BatchMatchFirst: Is / As will return true if errors.Is/errors.As passes on the
	// first error in Errs. Unwrap will return the first error in Errs if called
	// directly.
	//
	// BatchMatchAny: Is / As will return true if errors.Is/errors.As passes on ANY error
	// in Errs. Unwrap will return the first error in Errs if called directly.
	MatchMode BatchMatchMode
	// Errs are the OpError values we have collected.
	Errs []error
}

// Error implements builtins.error.
func (err BatchErrors) Error() string {
	return fmt.Sprintf(
		"%v errors returned. first: %v", len(err.Errs), err.Errs[0],
	)
}

// Unwrap implements xerrors.Wrapper.
func (err BatchErrors) Unwrap() error {
	// Panic if we do not contain any errors.
	if len(err.Errs) == 0 {
		panic("Unwrap() called on pears.BatchErrors value with no inner errors")
	}

	// Return an error based on the unwrap mode.
	switch err.MatchMode {
	case BatchMatchNone:
		return nil
	default:
		return err.Errs[0]
	}
}

// Is can be used by errors.Is to match on sub-errors.
func (err BatchErrors) Is(target error) bool {
	switch err.MatchMode {
	case BatchMatchNone:
		// We are not matching on sub-errors, return false.
		return false
	case BatchMatchAny:
		// Will return true if target passes errors.Is on ANY sub-errors.
		return err.matchAnyIs(target)
	default:
		// Otherwise we can call unwrap to handle the other modes, and compare the
		// result with errors.Is.
		compareTo := err.Unwrap()
		return errors.Is(compareTo, target)
	}
}

// matchAnyIs checks if ANY error in Errs matches target for errors.Is.
func (err BatchErrors) matchAnyIs(target error) bool {
	for _, thisErr := range err.Errs {
		if errors.Is(thisErr, target) {
			return true
		}
	}
	return false
}

// As can be used by errors.As to match on sub-errors.
func (err BatchErrors) As(target interface{}) bool {
	switch err.MatchMode {
	case BatchMatchNone:
		// We are not matching on sub-errors, return false.
		return false
	case BatchMatchAny:
		// Will return true if target passes errors.Is on ANY sub-errors.
		return err.matchAnyAs(target)
	default:
		// Otherwise we can call unwrap to handle the other modes, and compare the
		// result with errors.Is.
		compareTo := err.Unwrap()
		return errors.As(compareTo, target)
	}
}

// matchAnyAs checks if ANY error in Errs matches target for errors.As.
func (err BatchErrors) matchAnyAs(target interface{}) bool {
	for _, thisErr := range err.Errs {
		if errors.As(thisErr, target) {
			return true
		}
	}
	return false
}
