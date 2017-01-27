package main

import (
	"fmt"

	"github.com/reconquest/hierr-go"
)

var (
	_ hierr.HierarchicalError = (*Error)(nil)
)

// Error is smart and flexible string representation of occurred error, can be
// printed as plain one-line string or as hierarchical multi-leveled error.
type Error struct {
	Message string
	Nested  interface{}
}

// NewError returns instance of Error
func NewError(Nested interface{}, format string, args ...interface{}) Error {
	return Error{
		Message: fmt.Sprintf(format, args...),
		Nested:  Nested,
	}
}

// Error returns plain one-line string representation of occurred error, this
// method should be used for saving error to sould error logs.
func (err Error) Error() string {
	return err.Message + ": " + fmt.Sprintf("%s", err.Nested)
}

// HierarchicalError returns hierarchical (with unicode symbols) string
// representation of occurred error, this method used by hierr package for
// sending occurred slave errors to user as part of http response.
func (err Error) HierarchicalError() string {
	return hierr.Errorf(err.Nested, err.Message).Error()
}

// GetNested returns slice of nested errors.
func (err Error) GetNested() []hierr.NestedError {
	if sliced, ok := err.Nested.([]interface{}); ok {
		nesteds := []hierr.NestedError{}
		for _, nested := range sliced {
			nesteds = append(nesteds, nested)
		}

		return nesteds
	}

	return []hierr.NestedError{err.Nested}
}

// GetMessage returns top-level error message.
func (err Error) GetMessage() string {
	return err.Message
}
