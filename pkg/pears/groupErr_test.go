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
		// The GroupErrors value to test.
		Err pears.GroupErrors
		// The target error to test against.
		Target error
		// The expected result from errors.Is.
		IsExpected bool
	}{
		{
			Name: "None_HasMatch",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchNone,
				Errs: []error{
					io.EOF,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
		{
			Name: "None_NoMatch",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchNone,
				Errs: []error{
					io.ErrClosedPipe,
				},
			},
			Target:     io.EOF,
			IsExpected: false,
		},
		{
			Name: "Any_HasMatch",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
				Errs: []error{
					io.EOF,
				},
			},
			Target:     io.EOF,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_1stOf2",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
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
		// The GroupErrors value to test.
		Err pears.GroupErrors
		// The target error to test against.
		Target net.Error
		// The expected result from errors.Is.
		IsExpected bool
	}{
		{
			Name: "None_HasMatch",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchNone,
				Errs: []error{
					net.InvalidAddrError("mock error"),
				},
			},
			Target:     nil,
			IsExpected: false,
		},
		{
			Name: "None_NoMatch",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchNone,
				Errs: []error{
					io.ErrClosedPipe,
				},
			},
			Target:     nil,
			IsExpected: false,
		},
		{
			Name: "Any_HasMatch",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
				Errs: []error{
					net.InvalidAddrError("mock error"),
				},
			},
			Target:     nil,
			IsExpected: true,
		},
		{
			Name: "Any_HasMatch_1stOf2",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
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
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
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
		MatchMode pears.GroupMatchMode
	}{
		{
			Name:      "GroupMatchNone",
			MatchMode: pears.GroupMatchNone,
		},
		{
			Name:      "GroupMatchFirst",
			MatchMode: pears.GroupMatchFirst,
		},
		{
			Name:      "GroupMatchAny",
			MatchMode: pears.GroupMatchAny,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var err error = pears.GroupErrors{
				MatchMode: tc.MatchMode,
				Errs: []error{
					io.ErrClosedPipe,
					io.EOF,
					io.ErrClosedPipe,
				},
			}
			var batchErr pears.GroupErrors

			// Test that it works on a raw GroupErrors.
			t.Run("raw", func(t *testing.T) {
				assert.ErrorAs(t, err, &batchErr, "errors.As(*OpError)")
				// We'll use the Errs field count to verify we extracted our error.
				assert.Len(t, batchErr.Errs, 3, "batch err was extracted")
			})

			// Test that we can extract pears.GroupErrors from a wrapped error.
			err = fmt.Errorf("wrapping GroupErrors: %w", err)
			batchErr = pears.GroupErrors{}
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
		Err             pears.GroupErrors
		ExpectedMessage string
	}{
		{
			Name: "GroupMatchNone",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchNone,
				Errs: []error{
					io.ErrUnexpectedEOF,
					io.ErrNoProgress,
					io.EOF,
				},
			},
			ExpectedMessage: "3 errors returned. first: unexpected EOF",
		},
		{
			Name: "GroupMatchAny",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchFirst,
				Errs: []error{
					io.ErrUnexpectedEOF,
					io.ErrNoProgress,
					io.EOF,
				},
			},
			ExpectedMessage: "3 errors returned. first: unexpected EOF",
		},
		{
			Name: "GroupMatchAny",
			Err: pears.GroupErrors{
				MatchMode: pears.GroupMatchAny,
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
	var err error = pears.GroupErrors{}
	assert.Panics(t, func() {
		errors.Is(err, io.EOF)
	}, "unwrap on empty GroupErrors panics")
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
