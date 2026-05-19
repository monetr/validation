// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "reflect"

// ErrNotEqFieldInvalid is the error that returns when a value matches the value
// of another field.
var ErrNotEqFieldInvalid = NewError("validation_not_eq_field_invalid", "must not be equal to the other value")

// NotEqField returns a validation rule that checks if a value differs from the
// value pointed to by other. It is intended for comparing two struct fields,
// for example a new password that must not match the current one. Pass a
// pointer to the sibling field, the same way it is passed to Field:
//
//	validation.ValidateStruct(&s,
//	    validation.Field(&s.NewPassword, validation.NotEqField(&s.OldPassword)),
//	)
//
// reflect.DeepEqual() is used to determine if the two values are equal. An
// empty value is considered valid. Use the Required rule to make sure a value
// is not empty.
func NotEqField[T any](other *T) NotEqFieldRule[T] {
	return NotEqFieldRule[T]{
		other: other,
		err:   ErrNotEqFieldInvalid,
	}
}

// NotEqFieldRule is a validation rule that checks if a value differs from
// another field.
type NotEqFieldRule[T any] struct {
	other *T
	err   Error
}

// Validate checks if the given value is valid or not.
func (r NotEqFieldRule[T]) Validate(value any) error {
	value, isNil, err := Indirect(value)
	if err != nil {
		return err
	}
	if isNil || IsEmpty(value) {
		return nil
	}

	other, _, err := Indirect(*r.other)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(other, value) {
		return r.err
	}

	return nil
}

// Error sets the error message for the rule.
func (r NotEqFieldRule[T]) Error(message string) NotEqFieldRule[T] {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r NotEqFieldRule[T]) ErrorObject(err Error) NotEqFieldRule[T] {
	r.err = err
	return r
}
