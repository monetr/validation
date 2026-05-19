// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "reflect"

// ErrEqInvalid is the error that returns when a value is not equal to the
// expected value.
var ErrEqInvalid = NewError("validation_eq_invalid", "must be equal to {{.expected}}")

// Eq returns a validation rule that checks if a value is equal to the given
// value. reflect.DeepEqual() is used to determine if two values are equal. An
// empty value is considered valid. Use the Required rule to make sure a value
// is not empty.
func Eq[T any](expected T) EqRule[T] {
	return EqRule[T]{
		expected: expected,
		err:      ErrEqInvalid,
	}
}

// EqRule is a validation rule that checks if a value is equal to the expected
// value.
type EqRule[T any] struct {
	expected T
	err      Error
}

// Validate checks if the given value is valid or not.
func (r EqRule[T]) Validate(value any) error {
	value, isNil, err := Indirect(value)
	if err != nil {
		return err
	}
	if isNil || IsEmpty(value) {
		return nil
	}

	if reflect.DeepEqual(r.expected, value) {
		return nil
	}

	return r.err.SetParams(map[string]any{"expected": r.expected})
}

// Error sets the error message for the rule.
func (r EqRule[T]) Error(message string) EqRule[T] {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r EqRule[T]) ErrorObject(err Error) EqRule[T] {
	r.err = err
	return r
}
