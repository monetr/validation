// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

// Gt returns a validation rule that checks if a value is strictly greater than
// the specified value. This is equivalent to Min(min).Exclusive() and is
// provided as a more readable alternative.
// Note that the value being checked and the threshold value must be of the same
// type. Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure
// a value is not empty.
func Gt[T Threshold](min T) ThresholdRule[T] {
	return ThresholdRule[T]{
		threshold: min,
		operator:  greaterThan,
		err:       ErrMinGreaterThanRequired,
	}
}

// Gte returns a validation rule that checks if a value is greater than or equal
// to the specified value. This is equivalent to Min(min).
// Note that the value being checked and the threshold value must be of the same
// type. Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure
// a value is not empty.
func Gte[T Threshold](min T) ThresholdRule[T] {
	return ThresholdRule[T]{
		threshold: min,
		operator:  greaterEqualThan,
		err:       ErrMinGreaterEqualThanRequired,
	}
}

// Lt returns a validation rule that checks if a value is strictly less than the
// specified value. This is equivalent to Max(max).Exclusive() and is provided
// as a more readable alternative.
// Note that the value being checked and the threshold value must be of the same
// type. Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure
// a value is not empty.
func Lt[T Threshold](max T) ThresholdRule[T] {
	return ThresholdRule[T]{
		threshold: max,
		operator:  lessThan,
		err:       ErrMaxLessThanRequired,
	}
}

// Lte returns a validation rule that checks if a value is less than or equal to
// the specified value. This is equivalent to Max(max).
// Note that the value being checked and the threshold value must be of the same
// type. Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure
// a value is not empty.
func Lte[T Threshold](max T) ThresholdRule[T] {
	return ThresholdRule[T]{
		threshold: max,
		operator:  lessEqualThan,
		err:       ErrMaxLessEqualThanRequired,
	}
}
