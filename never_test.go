// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNever(t *testing.T) {
	str := "hello"
	empty := ""
	zero := 0
	nonZero := 5
	var nilStr *string

	tests := []struct {
		tag   string
		value any
		err   string
	}{
		{"untyped-nil", nil, ""},
		{"empty-string", "", ""},
		{"nonempty-string", "hello", "must not be provided"},
		{"zero-int", 0, ""},
		{"nonzero-int", 5, "must not be provided"},
		{"false-bool", false, ""},
		{"true-bool", true, "must not be provided"},
		{"empty-slice", []int{}, ""},
		{"nonempty-slice", []int{1}, "must not be provided"},
		{"empty-map", map[string]int{}, ""},
		{"nonempty-map", map[string]int{"a": 1}, "must not be provided"},
		{"nil-pointer", nilStr, ""},
		// A present pointer is a provided value and fails even when it references
		// an empty value: pointers get Nil semantics, not Empty semantics.
		{"pointer-to-nonempty", &str, "must not be provided"},
		{"pointer-to-empty-string", &empty, "must not be provided"},
		{"pointer-to-zero-int", &zero, "must not be provided"},
		{"pointer-to-nonzero-int", &nonZero, "must not be provided"},
	}

	for _, test := range tests {
		err := Never.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestNeverRule_Error(t *testing.T) {
	r := Never
	assert.Equal(t, "must not be provided", r.Validate("x").Error())

	r2 := r.Error("nope")
	// The original rule is unchanged (value receiver).
	assert.Equal(t, "must not be provided", r.Validate("x").Error())
	assert.Equal(t, "nope", r2.Validate("x").Error())
}

func TestNeverRule_ErrorObject(t *testing.T) {
	r := Never
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, "abc", r.Validate("x").Error())
	assert.NotEqual(t, err, Never.err)
}

// TestNever_InStructField exercises Never the way it is meant to be used: as a
// rule on a struct field that a variant forbids.
func TestNever_InStructField(t *testing.T) {
	type payload struct {
		Name  string  `json:"name"`
		Extra *string `json:"extra"`
	}

	// Non-pointer zero value and nil pointer both pass.
	p := payload{Name: "", Extra: nil}
	err := ValidateStruct(&p,
		Field(&p.Name, Never),
		Field(&p.Extra, Never),
	)
	assert.NoError(t, err)

	// A provided value on either field fails, keyed by the json tag.
	provided := "x"
	p = payload{Name: "supplied", Extra: &provided}
	err = ValidateStruct(&p,
		Field(&p.Name, Never),
		Field(&p.Extra, Never),
	)
	if assert.Error(t, err) {
		errs, ok := err.(Errors)
		assert.True(t, ok)
		assert.Equal(t, "must not be provided", errs["name"].Error())
		assert.Equal(t, "must not be provided", errs["extra"].Error())
	}
}
