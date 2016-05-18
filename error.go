package main

import (
	"fmt"

	"github.com/seletskiy/hierr"
)

// Error is smart and flexible string representation of occurred error, can be
// printed as plain one-line string or as hierarchical multi-leveled error.
type Error interface {
	error
	hierr.HierarchicalError
}

// FlexibleError implements Error interface.
type FlexibleError struct {
	format string
	args   []interface{}
	nested interface{}
}

// NewError returns instance of FlexibleError which implements Error interface.
func NewError(nested interface{}, format string, args ...interface{}) error {
	return FlexibleError{
		format: format,
		args:   args,
		nested: nested,
	}
}

// Error returns plain one-line string representation of occurred error, this
// method should be used for saving error to sould error logs.
func (err FlexibleError) Error() string {
	return fmt.Sprintf(err.format, err.args...) +
		": " + fmt.Sprintf("%s", err.nested)
}

// HierarchicalError returns hierarchical (with unicode symbols) string
// representation of occurred error, this method used by hierr package for
// sending occurred slave errors to user as part of http response.
func (err FlexibleError) HierarchicalError() string {
	return hierr.Errorf(err.nested, err.format, err.args...).Error()
}
