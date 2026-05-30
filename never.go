// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "reflect"

// ErrNever is the error returned by [Never] when a value that must be absent is
// instead present with a meaningful value.
var ErrNever = NewError("validation_never", "must not be provided")

// Never is a validation rule that asserts a value is absent. It is the
// discriminated-union counterpart to [Required]: use it on the struct fields
// that a given variant forbids.
//
// A struct field always exists (it is present as its zero value), so Never
// infers "absent" from the value itself:
//
//   - pointer or interface: must be nil. A non-nil pointer fails even when it
//     references an empty value, because a present pointer is a provided value.
//   - any other kind: must be the zero value (empty string, 0, false, empty
//     slice/map/array, the zero time.Time). A meaningful value fails.
//
// This auto-switching is what distinguishes Never from the existing absence
// rules: [Empty] dereferences first, so it would let a pointer to an empty
// value pass, while [Nil] would reject a non-pointer zero value. Never gives
// pointers Nil semantics and everything else Empty semantics in one rule.
//
// Never is value-based and intended for struct fields validated with [Field].
// It is deliberately not offered for map keys: a [Map] schema already reports
// keys it does not list as [ErrKeyUnexpected] (unless [MapRule.AllowExtraKeys]
// is set), so a map union forbids a key simply by omitting it from the variant
// that disallows it. Using Never as a map value rule would be a mistake — it
// would let a present-but-empty value pass when, in a map, the key's mere
// presence should fail.
var Never = neverRule{}

type neverRule struct {
	err Error
}

// Validate checks that the value is absent: a nil pointer/interface, or the
// zero value for any other kind.
func (r neverRule) Validate(value any) error {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Invalid:
		// An untyped nil is absent.
		return nil
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil
		}
	default:
		if IsEmpty(value) {
			return nil
		}
	}
	if r.err != nil {
		return r.err
	}
	return ErrNever
}

// Error sets the error message for the rule.
func (r neverRule) Error(message string) neverRule {
	if r.err == nil {
		r.err = ErrNever
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r neverRule) ErrorObject(err Error) neverRule {
	r.err = err
	return r
}
