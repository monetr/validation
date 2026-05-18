// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "reflect"

// ErrNotEqInvalid is the error that returns when a value is equal to the
// forbidden value.
var ErrNotEqInvalid = NewError("validation_not_eq_invalid", "must not be equal to {{.forbidden}}")

// NotEq returns a validation rule that checks if a value is not equal to the
// given value. reflect.DeepEqual() is used to determine if two values are
// equal. An empty value is considered valid. Use the Required rule to make sure
// a value is not empty.
func NotEq[T any](forbidden T) NotEqRule[T] {
	return NotEqRule[T]{
		forbidden: forbidden,
		err:       ErrNotEqInvalid,
	}
}

// NotEqRule is a validation rule that checks if a value is not equal to the
// forbidden value.
type NotEqRule[T any] struct {
	forbidden T
	err       Error
}

// Validate checks if the given value is valid or not.
func (r NotEqRule[T]) Validate(value any) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	if reflect.DeepEqual(r.forbidden, value) {
		return r.err.SetParams(map[string]any{"forbidden": r.forbidden})
	}

	return nil
}

// Error sets the error message for the rule.
func (r NotEqRule[T]) Error(message string) NotEqRule[T] {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r NotEqRule[T]) ErrorObject(err Error) NotEqRule[T] {
	r.err = err
	return r
}
