// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package is

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/monetr/validation"
	"github.com/stretchr/testify/assert"
)

// typeRuleCase is a single table entry: the rule under test, the value passed
// to Validate, and the expected error string ("" means the value is valid).
type typeRuleCase struct {
	tag      string
	rule     TypeRule
	value    any
	expected string
}

func runTypeRuleCases(t *testing.T, cases []typeRuleCase) {
	t.Helper()
	for _, c := range cases {
		err := c.rule.Validate(c.value)
		if c.expected == "" {
			assert.NoError(t, err, c.tag)
			continue
		}
		if assert.Error(t, err, c.tag) {
			assert.Equal(t, c.expected, err.Error(), c.tag)
		}
	}
}

func TestString(t *testing.T) {
	var nilPtr *string
	s := "present"

	runTypeRuleCases(t, []typeRuleCase{
		{"string", String, "abc", ""},
		{"empty string is still a string", String, "", ""},
		{"byte slice", String, []byte("abc"), ""},
		{"pointer to string", String, &s, ""},
		{"untyped nil is skipped", String, nil, ""},
		{"nil pointer is skipped", String, nilPtr, ""},

		// Present zero values of the wrong type are NOT skipped: a present 0 has
		// a type and that type is not string.
		{"present zero int", String, 0, "must be a string"},
		{"present false", String, false, "must be a string"},

		{"int", String, 5, "must be a string"},
		{"float", String, 5.5, "must be a string"},
		{"bool", String, true, "must be a string"},

		// A json.Number can only come from an unquoted JSON number token, so it
		// is provably never a string.
		{"json.Number is never a string", String, json.Number("5"), "must be a string"},
	})
}

func TestInteger(t *testing.T) {
	var nilPtr *int
	n := 7

	runTypeRuleCases(t, []typeRuleCase{
		{"int", Integer, 5, ""},
		{"present zero int", Integer, 0, ""},
		{"negative int", Integer, -5, ""},
		{"int8", Integer, int8(5), ""},
		{"int16", Integer, int16(5), ""},
		{"int32", Integer, int32(5), ""},
		{"int64", Integer, int64(5), ""},
		{"uint", Integer, uint(5), ""},
		{"uint64", Integer, uint64(5), ""},
		{"pointer to int", Integer, &n, ""},
		{"untyped nil is skipped", Integer, nil, ""},
		{"nil pointer is skipped", Integer, nilPtr, ""},

		// Whole-valued floats count: json.Unmarshal turns every number into a
		// float64, so 5 and 5.0 are indistinguishable.
		{"whole float", Integer, 5.0, ""},
		{"whole float32", Integer, float32(5.0), ""},
		{"negative whole float", Integer, -5.0, ""},

		// Fractional floats do not.
		{"fractional float", Integer, 5.5, "must be an integer"},
		{"NaN", Integer, math.NaN(), "must be an integer"},
		{"positive infinity", Integer, math.Inf(1), "must be an integer"},

		{"string is not an integer", Integer, "5", "must be an integer"},
		{"bool is not an integer", Integer, true, "must be an integer"},

		// json.Number classified by content.
		{"json integer", Integer, json.Number("5"), ""},
		{"json negative integer", Integer, json.Number("-5"), ""},
		{"json whole decimal", Integer, json.Number("5.0"), ""},
		{"json exponent whole", Integer, json.Number("1e2"), ""},
		{"json fractional", Integer, json.Number("5.5"), "must be an integer"},
		{"json non-numeric", Integer, json.Number("abc"), "must be an integer"},
		{"json empty", Integer, json.Number(""), "must be an integer"},
	})
}

func TestBoolean(t *testing.T) {
	var nilPtr *bool
	b := true

	runTypeRuleCases(t, []typeRuleCase{
		{"true", Boolean, true, ""},

		// false must NOT be skipped as an "empty" value: it is a present bool.
		{"false is still a boolean", Boolean, false, ""},
		{"pointer to bool", Boolean, &b, ""},
		{"untyped nil is skipped", Boolean, nil, ""},
		{"nil pointer is skipped", Boolean, nilPtr, ""},

		{"present zero int", Boolean, 0, "must be a boolean"},
		{"int", Boolean, 1, "must be a boolean"},
		{"string", Boolean, "true", "must be a boolean"},
		{"json.Number is never a boolean", Boolean, json.Number("1"), "must be a boolean"},
	})
}

func TestArray(t *testing.T) {
	var nilSlice []any
	var nilPtr *[]any
	present := []any{1, 2, 3}

	runTypeRuleCases(t, []typeRuleCase{
		{"slice of any", Array, []any{1, "two", true}, ""},
		{"empty but present slice is still an array", Array, []any{}, ""},
		{"typed slice", Array, []string{"a", "b"}, ""},
		{"fixed size array", Array, [3]int{1, 2, 3}, ""},
		{"pointer to slice", Array, &present, ""},

		{"untyped nil is skipped", Array, nil, ""},
		{"nil slice is skipped", Array, nilSlice, ""},
		{"nil pointer is skipped", Array, nilPtr, ""},

		// A []byte is string content as far as this library is concerned, so it
		// is NOT an array. This keeps String and Array from both claiming the
		// same value.
		{"byte slice is not an array", Array, []byte("abc"), "must be an array"},

		{"string is not an array", Array, "abc", "must be an array"},
		{"int is not an array", Array, 5, "must be an array"},
		{"map is not an array", Array, map[string]any{}, "must be an array"},
		{"json.Number is not an array", Array, json.Number("5"), "must be an array"},
	})
}

func TestMap(t *testing.T) {
	var nilMap map[string]any
	var nilPtr *map[string]any
	present := map[string]any{"a": 1}

	runTypeRuleCases(t, []typeRuleCase{
		{"map of string to any", Map, map[string]any{"a": 1, "b": "two"}, ""},
		{"empty but present map is still an object", Map, map[string]any{}, ""},
		{"map with non string keys", Map, map[int]string{1: "a"}, ""},
		{"pointer to map", Map, &present, ""},

		{"untyped nil is skipped", Map, nil, ""},
		{"nil map is skipped", Map, nilMap, ""},
		{"nil pointer is skipped", Map, nilPtr, ""},

		{"slice is not an object", Map, []any{1, 2}, "must be an object"},
		{"string is not an object", Map, "abc", "must be an object"},
		{"int is not an object", Map, 5, "must be an object"},
		{"json.Number is not an object", Map, json.Number("5"), "must be an object"},
	})
}

// TestTypeRule_WithRequired documents the intended composition: Required guards
// presence, the type rule guards the type.
func TestTypeRule_WithRequired(t *testing.T) {
	// A nil pointer fails Required but is skipped by String.
	var nilPtr *string
	assert.NoError(t, String.Validate(nilPtr))
	assert.Error(t, validation.Required.Validate(nilPtr))

	// A present value of the right type passes both.
	assert.NoError(t, validation.Validate("hello", validation.Required, String))

	// A present value of the wrong type is rejected by the type rule even
	// though Required is satisfied.
	assert.Equal(t, "must be a string", validation.Validate(5, validation.Required, String).Error())
}

// erroringValuer is a driver.Valuer whose Value() always fails, used to prove
// the rule surfaces the malfunction rather than silently treating it as absent.
type erroringValuer struct{}

func (erroringValuer) Value() (driver.Value, error) {
	return nil, errors.New("valuer boom")
}

// TestTypeRule_ValuerError ensures a malfunctioning driver.Valuer surfaces as
// an error rather than being silently treated as a valid (absent) value.
func TestTypeRule_ValuerError(t *testing.T) {
	err := String.Validate(erroringValuer{})
	if assert.Error(t, err) {
		ie, ok := err.(validation.InternalError)
		assert.True(t, ok, "expected an InternalError")
		assert.EqualError(t, ie.InternalError(), "valuer boom")
	}
}

func TestTypeRule_Error(t *testing.T) {
	r := Integer
	assert.Equal(t, "must be an integer", r.Validate("x").Error())

	r = r.Error("must be a whole number")
	assert.Equal(t, "must be a whole number", r.err.Message())
	assert.Equal(t, "must be a whole number", r.Validate("x").Error())

	// Customizing one rule must not mutate the shared package-level rule.
	assert.Equal(t, "must be an integer", Integer.Validate("x").Error())
}

func TestTypeRule_ErrorObject(t *testing.T) {
	r := String
	err := validation.NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
	assert.Equal(t, "abc", r.Validate(5).Error())
}
