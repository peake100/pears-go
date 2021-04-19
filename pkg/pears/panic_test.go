package pears_test

import (
	"errors"
	"github.com/peake100/pears-go/pkg/pears"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestCatchPanic_NoPanic(t *testing.T) {
	assert := assert.New(t)

	result := false
	err := pears.CatchPanic(func() (innerErr error) {
		result = true
		return nil
	})

	assert.NoError(err, "no error returned")
	assert.Equal(true, result, "result is true")
}

func TestCatchPanic_ErrReturned(t *testing.T) {
	assert := assert.New(t)

	err := pears.CatchPanic(func() (innerErr error) {
		return io.EOF
	})

	assert.Error(err, "error returned")
	assert.ErrorIs(err, io.EOF, "error is io.EOF")
}

func TestCatchPanic_Panic_ErrValue(t *testing.T) {
	assert := assert.New(t)

	err := pears.CatchPanic(func() (innerErr error) {
		panic(io.EOF)
	})

	assert.Error(err, "error returned")

	var panicErr pears.PanicError
	if !assert.ErrorAs(err, &panicErr, "error is panic error") {
		t.FailNow()
	}
	assert.ErrorIs(err, io.EOF, "error is io.EOF")
	assert.EqualError(err, "panic recovered: EOF")

	if !assert.NotNil(panicErr.Recovered, "PanicError.Recovered not nil") {
		t.FailNow()
	}
	if !assert.NotNil(panicErr.RecoveredErr, "PanicError.RecoveredErr not nil") {
		t.FailNow()
	}
	assert.ErrorIs(panicErr.RecoveredErr, io.EOF, "PanicError.RecoveredErr is io.EOF")
	assert.IsType(
		panicErr.Recovered,
		errors.New("some error"),
		"PanicError.Recovered is error string type",
	)
	assert.NotZero(panicErr.StackTrace, "stack trace not empty")
	t.Log("STACKTRACE:\n", panicErr.StackTrace)
}

func TestCatchPanic_Panic_IntValue(t *testing.T) {
	assert := assert.New(t)

	err := pears.CatchPanic(func() (innerErr error) {
		panic(2)
	})

	assert.Error(err, "error returned")

	var panicErr pears.PanicError
	if !assert.ErrorAs(err, &panicErr, "error is panic error") {
		t.FailNow()
	}
	assert.EqualError(err, "panic recovered: 2")

	if !assert.NotNil(panicErr.Recovered, "PanicError.Recovered not nil") {
		t.FailNow()
	}
	if !assert.NotNil(panicErr.RecoveredErr, "PanicError.RecoveredErr is nil") {
		t.FailNow()
	}
	assert.IsType(
		panicErr.Recovered,
		1,
		"PanicError.Recovered is error int type",
	)
	assert.IsType(
		panicErr.RecoveredErr,
		errors.New("mock"),
		"PanicError.Recovered is error int type",
	)
	t.Log("STACKTRACE:\n", panicErr.StackTrace)
}
