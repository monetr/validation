// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

// ErrBetweenInvalid is the error that returns when a value is not within the
// specified range.
var ErrBetweenInvalid = NewError("validation_between_invalid", "must be between {{.min}} and {{.max}}")

// Between returns a validation rule that checks if a value is within the
// inclusive range [min, max]. By calling Exclusive, both boundaries are
// excluded from the accepted range.
// Note that the value being checked and the boundary values must be of the same
// type. Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure
// a value is not empty.
func Between[T Threshold](min, max T) BetweenRule[T] {
	return BetweenRule[T]{
		min: min,
		max: max,
		err: ErrBetweenInvalid,
	}
}

// BetweenRule is a validation rule that checks if a value is within a range.
type BetweenRule[T Threshold] struct {
	min       T
	max       T
	exclusive bool
	err       Error
}

// Exclusive sets the comparison to exclude both boundary values.
func (r BetweenRule[T]) Exclusive() BetweenRule[T] {
	r.exclusive = true
	return r
}

// Validate checks if the given value is within the range.
func (r BetweenRule[T]) Validate(value any) error {
	lower := Gte(r.min)
	upper := Lte(r.max)
	if r.exclusive {
		lower = Gt(r.min)
		upper = Lt(r.max)
	}

	for _, rule := range []ThresholdRule[T]{lower, upper} {
		err := rule.Validate(value)
		if err == nil {
			continue
		}
		// A threshold violation surfaces as a package Error. Anything else (a
		// type or conversion failure) is unrelated to the range and is returned
		// untouched.
		if _, ok := err.(Error); !ok {
			return err
		}
		return r.err.SetParams(map[string]any{"min": r.min, "max": r.max})
	}

	return nil
}

// Error sets the error message for the rule.
func (r BetweenRule[T]) Error(message string) BetweenRule[T] {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r BetweenRule[T]) ErrorObject(err Error) BetweenRule[T] {
	r.err = err
	return r
}
