// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

var (
	// ErrMinGreaterEqualThanRequired is the error that returns when a value is less than a specified threshold.
	ErrMinGreaterEqualThanRequired = NewError("validation_min_greater_equal_than_required", "must be no less than {{.threshold}}")
	// ErrMaxLessEqualThanRequired is the error that returns when a value is greater than a specified threshold.
	ErrMaxLessEqualThanRequired = NewError("validation_max_less_equal_than_required", "must be no greater than {{.threshold}}")
	// ErrMinGreaterThanRequired is the error that returns when a value is less than or equal to a specified threshold.
	ErrMinGreaterThanRequired = NewError("validation_min_greater_than_required", "must be greater than {{.threshold}}")
	// ErrMaxLessThanRequired is the error that returns when a value is greater than or equal to a specified threshold.
	ErrMaxLessThanRequired = NewError("validation_max_less_than_required", "must be less than {{.threshold}}")
)

// ThresholdRule is a validation rule that checks if a value satisfies the specified threshold requirement.
type ThresholdRule struct {
	threshold any
	operator  int
	err       Error
}

const (
	greaterThan = iota
	greaterEqualThan
	lessThan
	lessEqualThan
)

// Min returns a validation rule that checks if a value is greater or equal than the specified value.
// By calling Exclusive, the rule will check if the value is strictly greater than the specified value.
// Note that the value being checked and the threshold value must be of the same type.
// Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure a value is not empty.
func Min(min any) ThresholdRule {
	return ThresholdRule{
		threshold: min,
		operator:  greaterEqualThan,
		err:       ErrMinGreaterEqualThanRequired,
	}

}

// Max returns a validation rule that checks if a value is less or equal than the specified value.
// By calling Exclusive, the rule will check if the value is strictly less than the specified value.
// Note that the value being checked and the threshold value must be of the same type.
// Only int, uint, float and time.Time types are supported.
// An empty value is considered valid. Please use the Required rule to make sure a value is not empty.
func Max(max any) ThresholdRule {
	return ThresholdRule{
		threshold: max,
		operator:  lessEqualThan,
		err:       ErrMaxLessEqualThanRequired,
	}
}

// Exclusive sets the comparison to exclude the boundary value.
func (r ThresholdRule) Exclusive() ThresholdRule {
	switch r.operator {
	case greaterEqualThan:
		r.operator = greaterThan
		r.err = ErrMinGreaterThanRequired
	case lessEqualThan:
		r.operator = lessThan
		r.err = ErrMaxLessThanRequired
	}
	return r
}

// Validate checks if the given value is valid or not.
func (r ThresholdRule) Validate(value any) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return nil
	}

	if jsonNumber, ok := value.(json.Number); ok {
		switch r.threshold.(type) {
		case int, int8, int16, int32, int64:
			// If our comparing number is an integer, then parse the json number as an
			// integer.
			i, err := jsonNumber.Int64()
			if err != nil {
				return err
			}
			value = i
		case uint, uint8, uint16, uint32, uint64:
			// If our comparing number is an unsigned integer, then parse the json
			// number as an integer and cast it.
			i, err := jsonNumber.Int64()
			if err != nil {
				return err
			}
			value = uint64(i)
		case float32, float64:
			// If our comparing value is a float, then parse the json number as a
			// float.
			f, err := jsonNumber.Float64()
			if err != nil {
				return err
			}
			value = f
		case time.Time:
			// If the value we have is a json number but we are comparing it to a time
			// object. Then assume it is a unix timestamp as a number.
			i, err := jsonNumber.Int64()
			if err != nil {
				return err
			}
			value = time.Unix(i, 0)
		}

		if IsEmpty(value) {
			return nil
		}
	}

	rv := reflect.ValueOf(r.threshold)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := ToInt(value)
		if err != nil {
			return err
		}
		if r.compareInt(rv.Int(), v) {
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := ToUint(value)
		if err != nil {
			return err
		}
		if r.compareUint(rv.Uint(), v) {
			return nil
		}

	case reflect.Float32, reflect.Float64:
		v, err := ToFloat(value)
		if err != nil {
			return err
		}
		if r.compareFloat(rv.Float(), v) {
			return nil
		}

	case reflect.Struct:
		t, ok := r.threshold.(time.Time)
		if !ok {
			return fmt.Errorf("type not supported: %v", rv.Type())
		}
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("cannot convert %v to time.Time", reflect.TypeOf(value))
		}
		if v.IsZero() || r.compareTime(t, v) {
			return nil
		}

	default:
		return fmt.Errorf("type not supported: %v", rv.Type())
	}

	return r.err.SetParams(map[string]any{"threshold": r.threshold})
}

// Error sets the error message for the rule.
func (r ThresholdRule) Error(message string) ThresholdRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r ThresholdRule) ErrorObject(err Error) ThresholdRule {
	r.err = err
	return r
}

func (r ThresholdRule) compareInt(threshold, value int64) bool {
	switch r.operator {
	case greaterThan:
		return value > threshold
	case greaterEqualThan:
		return value >= threshold
	case lessThan:
		return value < threshold
	default:
		return value <= threshold
	}
}

func (r ThresholdRule) compareUint(threshold, value uint64) bool {
	switch r.operator {
	case greaterThan:
		return value > threshold
	case greaterEqualThan:
		return value >= threshold
	case lessThan:
		return value < threshold
	default:
		return value <= threshold
	}
}

func (r ThresholdRule) compareFloat(threshold, value float64) bool {
	switch r.operator {
	case greaterThan:
		return value > threshold
	case greaterEqualThan:
		return value >= threshold
	case lessThan:
		return value < threshold
	default:
		return value <= threshold
	}
}

func (r ThresholdRule) compareTime(threshold, value time.Time) bool {
	switch r.operator {
	case greaterThan:
		return value.After(threshold)
	case greaterEqualThan:
		return value.After(threshold) || value.Equal(threshold)
	case lessThan:
		return value.Before(threshold)
	default:
		return value.Before(threshold) || value.Equal(threshold)
	}
}
