// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"encoding/json"
	"math"
	"testing"

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

func TestIsString(t *testing.T) {
	var nilPtr *string
	s := "present"

	runTypeRuleCases(t, []typeRuleCase{
		{"string", IsString, "abc", ""},
		{"empty string is still a string", IsString, "", ""},
		{"byte slice", IsString, []byte("abc"), ""},
		{"pointer to string", IsString, &s, ""},
		{"untyped nil is skipped", IsString, nil, ""},
		{"nil pointer is skipped", IsString, nilPtr, ""},

		// Present zero values of the wrong type are NOT skipped: a present 0 has
		// a type and that type is not string.
		{"present zero int", IsString, 0, "must be a string"},
		{"present false", IsString, false, "must be a string"},

		{"int", IsString, 5, "must be a string"},
		{"float", IsString, 5.5, "must be a string"},
		{"bool", IsString, true, "must be a string"},

		// A json.Number can only come from an unquoted JSON number token, so it
		// is provably never a string.
		{"json.Number is never a string", IsString, json.Number("5"), "must be a string"},
	})
}

func TestIsInteger(t *testing.T) {
	var nilPtr *int
	n := 7

	runTypeRuleCases(t, []typeRuleCase{
		{"int", IsInteger, 5, ""},
		{"present zero int", IsInteger, 0, ""},
		{"negative int", IsInteger, -5, ""},
		{"int8", IsInteger, int8(5), ""},
		{"int16", IsInteger, int16(5), ""},
		{"int32", IsInteger, int32(5), ""},
		{"int64", IsInteger, int64(5), ""},
		{"uint", IsInteger, uint(5), ""},
		{"uint64", IsInteger, uint64(5), ""},
		{"pointer to int", IsInteger, &n, ""},
		{"untyped nil is skipped", IsInteger, nil, ""},
		{"nil pointer is skipped", IsInteger, nilPtr, ""},

		// Whole-valued floats count: json.Unmarshal turns every number into a
		// float64, so 5 and 5.0 are indistinguishable.
		{"whole float", IsInteger, 5.0, ""},
		{"whole float32", IsInteger, float32(5.0), ""},
		{"negative whole float", IsInteger, -5.0, ""},

		// Fractional floats do not.
		{"fractional float", IsInteger, 5.5, "must be an integer"},
		{"NaN", IsInteger, math.NaN(), "must be an integer"},
		{"positive infinity", IsInteger, math.Inf(1), "must be an integer"},

		{"string is not an integer", IsInteger, "5", "must be an integer"},
		{"bool is not an integer", IsInteger, true, "must be an integer"},

		// json.Number classified by content.
		{"json integer", IsInteger, json.Number("5"), ""},
		{"json negative integer", IsInteger, json.Number("-5"), ""},
		{"json whole decimal", IsInteger, json.Number("5.0"), ""},
		{"json exponent whole", IsInteger, json.Number("1e2"), ""},
		{"json fractional", IsInteger, json.Number("5.5"), "must be an integer"},
		{"json non-numeric", IsInteger, json.Number("abc"), "must be an integer"},
		{"json empty", IsInteger, json.Number(""), "must be an integer"},
	})
}

func TestIsFloat(t *testing.T) {
	var nilPtr *float64
	f := 1.5

	runTypeRuleCases(t, []typeRuleCase{
		{"float", IsFloat, 5.5, ""},
		{"float32", IsFloat, float32(5.5), ""},
		{"present zero float", IsFloat, 0.0, ""},
		{"pointer to float", IsFloat, &f, ""},
		{"untyped nil is skipped", IsFloat, nil, ""},
		{"nil pointer is skipped", IsFloat, nilPtr, ""},

		// Every integer is a valid number.
		{"int is a number", IsFloat, 5, ""},
		{"present zero int is a number", IsFloat, 0, ""},
		{"uint is a number", IsFloat, uint(5), ""},

		{"string is not a number", IsFloat, "5.5", "must be a number"},
		{"bool is not a number", IsFloat, true, "must be a number"},

		// json.Number classified by content.
		{"json fractional", IsFloat, json.Number("5.5"), ""},
		{"json integer is a number", IsFloat, json.Number("5"), ""},
		{"json exponent", IsFloat, json.Number("1.5e3"), ""},
		{"json non-numeric", IsFloat, json.Number("abc"), "must be a number"},
		{"json empty", IsFloat, json.Number(""), "must be a number"},
	})
}

func TestIsBoolean(t *testing.T) {
	var nilPtr *bool
	b := true

	runTypeRuleCases(t, []typeRuleCase{
		{"true", IsBoolean, true, ""},

		// false must NOT be skipped as an "empty" value: it is a present bool.
		{"false is still a boolean", IsBoolean, false, ""},
		{"pointer to bool", IsBoolean, &b, ""},
		{"untyped nil is skipped", IsBoolean, nil, ""},
		{"nil pointer is skipped", IsBoolean, nilPtr, ""},

		{"present zero int", IsBoolean, 0, "must be a boolean"},
		{"int", IsBoolean, 1, "must be a boolean"},
		{"string", IsBoolean, "true", "must be a boolean"},
		{"json.Number is never a boolean", IsBoolean, json.Number("1"), "must be a boolean"},
	})
}

func TestIsArray(t *testing.T) {
	var nilSlice []any
	var nilPtr *[]any
	present := []any{1, 2, 3}

	runTypeRuleCases(t, []typeRuleCase{
		{"slice of any", IsArray, []any{1, "two", true}, ""},
		{"empty but present slice is still an array", IsArray, []any{}, ""},
		{"typed slice", IsArray, []string{"a", "b"}, ""},
		{"fixed size array", IsArray, [3]int{1, 2, 3}, ""},
		{"pointer to slice", IsArray, &present, ""},

		{"untyped nil is skipped", IsArray, nil, ""},
		{"nil slice is skipped", IsArray, nilSlice, ""},
		{"nil pointer is skipped", IsArray, nilPtr, ""},

		// A []byte is string content as far as this library is concerned, so it
		// is NOT an array. This keeps IsString and IsArray from both claiming the
		// same value.
		{"byte slice is not an array", IsArray, []byte("abc"), "must be an array"},

		{"string is not an array", IsArray, "abc", "must be an array"},
		{"int is not an array", IsArray, 5, "must be an array"},
		{"map is not an array", IsArray, map[string]any{}, "must be an array"},
		{"json.Number is not an array", IsArray, json.Number("5"), "must be an array"},
	})
}

func TestIsMap(t *testing.T) {
	var nilMap map[string]any
	var nilPtr *map[string]any
	present := map[string]any{"a": 1}

	runTypeRuleCases(t, []typeRuleCase{
		{"map of string to any", IsMap, map[string]any{"a": 1, "b": "two"}, ""},
		{"empty but present map is still an object", IsMap, map[string]any{}, ""},
		{"map with non string keys", IsMap, map[int]string{1: "a"}, ""},
		{"pointer to map", IsMap, &present, ""},

		{"untyped nil is skipped", IsMap, nil, ""},
		{"nil map is skipped", IsMap, nilMap, ""},
		{"nil pointer is skipped", IsMap, nilPtr, ""},

		{"slice is not an object", IsMap, []any{1, 2}, "must be an object"},
		{"string is not an object", IsMap, "abc", "must be an object"},
		{"int is not an object", IsMap, 5, "must be an object"},
		{"json.Number is not an object", IsMap, json.Number("5"), "must be an object"},
	})
}

// TestTypeRule_WithRequired documents the intended composition: Required guards
// presence, the type rule guards the type.
func TestTypeRule_WithRequired(t *testing.T) {
	// A nil pointer fails Required but is skipped by IsString.
	var nilPtr *string
	assert.NoError(t, IsString.Validate(nilPtr))
	assert.Error(t, Required.Validate(nilPtr))

	// A present value of the right type passes both.
	assert.NoError(t, Validate("hello", Required, IsString))

	// A present value of the wrong type is rejected by the type rule even
	// though Required is satisfied.
	assert.Equal(t, "must be a string", Validate(5, Required, IsString).Error())
}

// TestTypeRule_ValuerError ensures a malfunctioning driver.Valuer surfaces as
// an error rather than being silently treated as a valid (absent) value.
func TestTypeRule_ValuerError(t *testing.T) {
	err := IsString.Validate(erroringValuer{})
	if assert.Error(t, err) {
		ie, ok := err.(InternalError)
		assert.True(t, ok, "expected an InternalError")
		assert.EqualError(t, ie.InternalError(), "valuer boom")
	}
}

func TestTypeRule_Error(t *testing.T) {
	r := IsInteger
	assert.Equal(t, "must be an integer", r.Validate("x").Error())

	r = r.Error("must be a whole number")
	assert.Equal(t, "must be a whole number", r.err.Message())
	assert.Equal(t, "must be a whole number", r.Validate("x").Error())

	// Customizing one rule must not mutate the shared package-level rule.
	assert.Equal(t, "must be an integer", IsInteger.Validate("x").Error())
}

func TestTypeRule_ErrorObject(t *testing.T) {
	r := IsString
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
	assert.Equal(t, "abc", r.Validate(5).Error())
}
