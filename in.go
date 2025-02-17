// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"reflect"
)

// ErrInInvalid is the error that returns in case of an invalid value for "in"
// rule.
var ErrInInvalid = NewError("validation_in_invalid", "must be a valid value")

// In returns a validation rule that checks if a value can be found in the given
// list of values. reflect.DeepEqual() will be used to determine if two values
// are equal. For more details please refer to
// https://golang.org/pkg/reflect/#DeepEqual An empty value is considered valid.
// Use the Required rule to make sure a value is not empty.
func In[T any](values ...T) InRule[T] {
	return InRule[T]{
		elements: values,
		err:      ErrInInvalid,
	}
}

// InRule is a validation rule that validates if a value can be found in the
// given list of values.
type InRule[T any] struct {
	elements []T
	err      Error
}

// Validate checks if the given value is valid or not.
func (r InRule[T]) Validate(value any) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	for _, e := range r.elements {
		if reflect.DeepEqual(e, value) {
			return nil
		}
	}

	return r.err
}

// Error sets the error message for the rule.
func (r InRule[T]) Error(message string) InRule[T] {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r InRule[T]) ErrorObject(err Error) InRule[T] {
	r.err = err
	return r
}
