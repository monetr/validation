// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "reflect"

// ErrEqFieldInvalid is the error that returns when a value does not match the
// value of another field.
var ErrEqFieldInvalid = NewError("validation_eq_field_invalid", "must be equal to the other value")

// EqField returns a validation rule that checks if a value is equal to the
// value pointed to by other. It is intended for comparing two struct fields,
// for example a password and its confirmation. Pass a pointer to the sibling
// field, the same way it is passed to Field:
//
//	validation.ValidateStruct(&s,
//	    validation.Field(&s.Password, validation.Required),
//	    validation.Field(&s.ConfirmPassword, validation.EqField(&s.Password)),
//	)
//
// reflect.DeepEqual() is used to determine if the two values are equal. An
// empty value is considered valid. Use the Required rule to make sure a value
// is not empty.
func EqField[T any](other *T) EqFieldRule[T] {
	return EqFieldRule[T]{
		other: other,
		err:   ErrEqFieldInvalid,
	}
}

// EqFieldRule is a validation rule that checks if a value equals another field.
type EqFieldRule[T any] struct {
	other *T
	err   Error
}

// Validate checks if the given value is valid or not.
func (r EqFieldRule[T]) Validate(value any) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	other, _ := Indirect(*r.other)
	if reflect.DeepEqual(other, value) {
		return nil
	}

	return r.err
}

// Error sets the error message for the rule.
func (r EqFieldRule[T]) Error(message string) EqFieldRule[T] {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r EqFieldRule[T]) ErrorObject(err Error) EqFieldRule[T] {
	r.err = err
	return r
}
