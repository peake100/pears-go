package pears_test

import (
	"errors"
	"fmt"
	"github.com/peake100/pears-go/pkg/pears"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"testing"
)

func TestBatchErrors_Is(t *testing.T) {
	testCases := []struct {
		// The name of the test case.
		Name string
		// The BatchErrors value to test.
		Err pears.BatchErrors
		// The target error to test against.
		Target error
		// The expected result from errors.Is.
		IsExpected bool
	}{
		{
			Name: "None_HasMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchNone,
				Errs: []error{
					io.EOF,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
		{
			Name: "None_NoMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchNone,
				Errs: []error{
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
		{
			Name: "Any_HasMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.EOF,
				},
			},
			Target:     io.EOF,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_1stOf2",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.EOF,
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_2ndOf2",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrClosedPipe,
					io.EOF,
				},
			},
			Target:     io.EOF,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_2ndOf3",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrClosedPipe,
					io.EOF,
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: true,
		},
		{
			Name: "Any_NoMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrClosedPipe,
					io.ErrClosedPipe,
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
		{
			Name: "Fist_HasMatch_First",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					io.EOF,
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: true,
		},
		{
			Name: "Fist_HasMatch_Second",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					io.ErrClosedPipe,
					io.EOF,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
		{
			Name: "Fist_NoMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					io.ErrClosedPipe,
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("errors.Is", func(t *testing.T) {
				result := errors.Is(tc.Err, tc.Target)
				assert.Equal(t, tc.IsExpected, result, "%v is %v", tc.Err, tc.Target)
			})
		})
	}
}

func TestBatchErrors_As(t *testing.T) {
	testCases := []struct {
		// The name of the test case.
		Name string
		// The BatchErrors value to test.
		Err pears.BatchErrors
		// The target error to test against.
		Target net.Error
		// The expected result from errors.Is.
		IsExpected bool
	}{
		{
			Name: "None_HasMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchNone,
				Errs: []error{
					net.InvalidAddrError("mock error"),
				},
			},
			Target:     nil,
			IsExpected: false,
		},
		{
			Name: "None_NoMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchNone,
				Errs: []error{
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: false,
		},
		{
			Name: "Any_HasMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					net.InvalidAddrError("mock error"),
				},
			},
			Target:     nil,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_1stOf2",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					net.InvalidAddrError("mock error"),
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_2ndOf2",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrClosedPipe,
					net.InvalidAddrError("mock error"),
				},
			},
			Target:     nil,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_2ndOf3",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrClosedPipe,
					net.InvalidAddrError("mock error"),
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: true,
		},
		{
			Name: "Any_NoMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrClosedPipe,
					io.ErrClosedPipe,
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: false,
		},
		{
			Name: "Fist_HasMatch_First",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					net.InvalidAddrError("mock error"),
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: true,
		},
		{
			Name: "Fist_HasMatch_Second",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					io.ErrClosedPipe,
					net.InvalidAddrError("mock error"),
				},
			},
			Target:     nil,
			IsExpected: false,
		},
		{
			Name: "Fist_NoMatch",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					io.ErrClosedPipe,
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Run("errors.Is", func(t *testing.T) {
				result := errors.As(tc.Err, &tc.Target)
				assert.Equal(t, tc.IsExpected, result, "%v is %v", tc.Err, tc.Target)
				if !tc.IsExpected {
					return
				}
				assert.EqualError(t, tc.Target, "mock error")
			})
		})
	}
}

func TestBatchErrors_As_MatchBatchError(t *testing.T) {
	testCases := []struct {
		Name      string
		MatchMode pears.BatchMatchMode
	}{
		{
			Name:      "BatchMatchNone",
			MatchMode: pears.BatchMatchNone,
		},
		{
			Name:      "BatchMatchFirst",
			MatchMode: pears.BatchMatchFirst,
		},
		{
			Name:      "BatchMatchAny",
			MatchMode: pears.BatchMatchAny,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var err error = pears.BatchErrors{
				MatchMode: tc.MatchMode,
				Errs: []error{
					io.ErrClosedPipe,
					io.EOF,
					io.ErrClosedPipe,
				},
			}
			var batchErr pears.BatchErrors

			// Test that it works on a raw BatchErrors.
			t.Run("raw", func(t *testing.T) {
				assert.ErrorAs(t, err, &batchErr, "errors.As(*OpError)")
				// We'll use the Errs field count to verify we extracted our error.
				assert.Len(t, batchErr.Errs, 3, "batch err was extracted")
			})

			// Test that we can extract pears.BatchErrors from a wrapped error.
			err = fmt.Errorf("wrapping BatchErrors: %w", err)
			batchErr = pears.BatchErrors{}
			t.Run("wrapped", func(t *testing.T) {
				assert.ErrorAs(t, err, &batchErr, "errors.As(*OpError)")
				// We'll use the Errs field count to verify we extracted our error.
				assert.Len(t, batchErr.Errs, 3, "batch err was extracted")
			})
		})
	}
}

func TestBatchErrors_Error(t *testing.T) {
	testCases := []struct {
		Name            string
		Err             pears.BatchErrors
		ExpectedMessage string
	}{
		{
			Name: "BatchMatchNone",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchNone,
				Errs: []error{
					io.ErrUnexpectedEOF,
					io.ErrNoProgress,
					io.EOF,
				},
			},
			ExpectedMessage: "3 errors returned. first: unexpected EOF",
		},
		{
			Name: "BatchMatchAny",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchFirst,
				Errs: []error{
					io.ErrUnexpectedEOF,
					io.ErrNoProgress,
					io.EOF,
				},
			},
			ExpectedMessage: "3 errors returned. first: unexpected EOF",
		},
		{
			Name: "BatchMatchAny",
			Err: pears.BatchErrors{
				MatchMode: pears.BatchMatchAny,
				Errs: []error{
					io.ErrUnexpectedEOF,
					io.ErrNoProgress,
					io.EOF,
				},
			},
			ExpectedMessage: "3 errors returned. first: unexpected EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.EqualError(t, tc.Err, tc.ExpectedMessage, "error text expected")
		})
	}
}

func TestBatchErrors_Unwrap_Panic(t *testing.T) {
	var err error = pears.BatchErrors{}
	assert.Panics(t, func() {
		errors.Is(err, io.EOF)
	}, "unwrap on empty BatchErrors panics")
}

func TestBatchError(t *testing.T) {
	err := pears.OpError{
		OpName: "read file",
		Err:    io.EOF,
	}

	t.Run("Error", func(t *testing.T) {
		assert.EqualError(t, err, "error during 'read file': EOF")
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.ErrorIs(t, err, io.EOF, "error unwraps to io.EOF")
	})
}
