// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package is

import (
	"encoding/json"
	"math"
	"reflect"

	"github.com/monetr/validation"
)

// bytesType is used to tell a []byte apart from other slices. The validation
// package has its own copy of this, but its unexported over there so we keep our
// own here rather than reaching across the package boundary for it.
var bytesType = reflect.TypeOf([]byte(nil))

// typeKind enumerates the categories asserted by the type rules.
type typeKind int

const (
	typeString typeKind = iota
	typeInteger
	typeFloat
	typeBool
	typeArray
	typeMap
)

var (
	// ErrTypeString is the error returned when a value is not a string.
	ErrTypeString = validation.NewError("validation_type_string", "must be a string")
	// ErrTypeInteger is the error returned when a value is not an integer.
	ErrTypeInteger = validation.NewError("validation_type_integer", "must be an integer")
	// ErrTypeFloat is the error returned when a value is not a number.
	ErrTypeFloat = validation.NewError("validation_type_float", "must be a number")
	// ErrTypeBool is the error returned when a value is not a boolean.
	ErrTypeBool = validation.NewError("validation_type_bool", "must be a boolean")
	// ErrTypeArray is the error returned when a value is not an array.
	ErrTypeArray = validation.NewError("validation_type_array", "must be an array")
	// ErrTypeMap is the error returned when a value is not an object. The JSON
	// term is used in the message because these rules are meant for JSON decoded
	// data, where a Go map is the object.
	ErrTypeMap = validation.NewError("validation_type_map", "must be an object")
)

// String, Integer, Boolean, Array, and Map assert the underlying type of a
// value. They are intended for values whose static type is dynamic (any) — for
// example data decoded from JSON into an interface or a map[string]any — where
// the Go compiler can no longer guarantee the type. When the static type is
// already concrete there is nothing for these rules to check.
//
// Unlike the value-oriented rules (Length, Match, Min, ...) which dereference
// to an empty value and treat it as valid, the type rules only skip a nil
// pointer/interface. A present zero value such as 0, "", or false carries a
// type, so the rule still asserts it: String rejects a present 0, while
// Integer accepts a present 0. Use Required to additionally demand presence.
//
// Numbers are matched by value, not by Go kind, so they behave the same whether
// JSON was decoded with the default float64 numbers or with json.Decoder's
// UseNumber:
//
//   - Integer accepts any integer-valued number, including a whole-valued
//     float such as 5.0 (but not 5.5). This is required because the default
//     json.Unmarshal turns every number into a float64, so 5 and 5.0 are
//     indistinguishable; Integer asserts the value, not how it was spelled.
//
// A json.Number is classified by its textual content for Integer. It can only
// originate from an unquoted JSON number token, never from a quoted string, so
// it is never a string (nor a boolean): String and Boolean always reject a
// json.Number.
//
// Array and Map assert the two structural JSON types. A JSON array decodes to a
// slice (the default []any) and a JSON object decodes to a map (the default
// map[string]any), so those are what these rules accept. Array deliberately
// does NOT accept a []byte: the library treats a byte slice as string content
// everywhere else (see String), so a []byte is a string here, not an array.
// Like the other type rules they only skip a true nil, which for these includes
// a nil slice/map, a present but empty []any{} or map[string]any{} still carries
// its type and so is accepted.
//
// Note that there is intentionally no is.Float type rule here. The is package
// already has an is.Float that checks whether a string CONTAINS a floating point
// number, which is a totally different thing from asserting that a value IS a
// number. Rather than overload the name and confuse everyone, the numeric type
// rule is left as validation.IsFloat for now.
var (
	String  = TypeRule{kind: typeString, err: ErrTypeString}
	Integer = TypeRule{kind: typeInteger, err: ErrTypeInteger}
	Boolean = TypeRule{kind: typeBool, err: ErrTypeBool}
	Array   = TypeRule{kind: typeArray, err: ErrTypeArray}
	Map     = TypeRule{kind: typeMap, err: ErrTypeMap}
)

// TypeRule is a validation rule that asserts the underlying type of a value.
// Use the package-level String, Integer, Boolean, Array, and Map rules rather
// than constructing one directly.
type TypeRule struct {
	kind typeKind
	err  validation.Error
}

// Error sets the error message for the rule.
func (r TypeRule) Error(message string) TypeRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r TypeRule) ErrorObject(err validation.Error) TypeRule {
	r.err = err
	return r
}

// Validate checks that the value's underlying type matches the asserted type.
func (r TypeRule) Validate(value any) error {
	value, isNil, err := validation.Indirect(value)
	if err != nil {
		return err
	}
	// Only a nil pointer/interface is skipped; a present zero value still
	// carries a type and is therefore asserted. See the String doc comment.
	if isNil {
		return nil
	}

	if r.matches(value) {
		return nil
	}
	return r.err
}

// matches reports whether value satisfies the rule's asserted type.
func (r TypeRule) matches(value any) bool {
	// json.Number's underlying kind is string, so it must be classified by its
	// textual content before any kind-based check below would misread it.
	if n, ok := value.(json.Number); ok {
		switch r.kind {
		case typeInteger:
			return isWholeJSONNumber(n)
		case typeFloat:
			_, err := n.Float64()
			return err == nil
		default:
			// A json.Number is provably not a string or boolean: it can only be
			// produced from an unquoted JSON number token.
			return false
		}
	}

	switch r.kind {
	case typeString:
		// EnsureString also accepts a []byte, matching how StringRule treats
		// string content throughout the library. A json.Number (kind string) is
		// already handled above and never reaches here.
		_, err := validation.EnsureString(value)
		return err == nil
	case typeInteger:
		return isIntegerValue(value)
	case typeFloat:
		return isNumericValue(value)
	case typeBool:
		return reflect.ValueOf(value).Kind() == reflect.Bool
	case typeArray:
		return isArrayValue(value)
	case typeMap:
		return reflect.ValueOf(value).Kind() == reflect.Map
	}
	return false
}

// isArrayValue reports whether value is a JSON style array: a slice or an array.
// A []byte is excluded on purpose. EnsureString treats a byte slice as string
// content, so String already claims it, and we dont want the same value to be
// both a string and an array. A byte array like [4]byte is not what EnsureString
// matches (it only matches the []byte slice type) so it stays an array here.
func isArrayValue(value any) bool {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice:
		return rv.Type() != bytesType
	case reflect.Array:
		return true
	}
	return false
}

// isIntegerValue reports whether value is an integer-valued number: any signed
// or unsigned integer type, or a float with no fractional part (so 5.0 counts,
// 5.5 does not).
func isIntegerValue(value any) bool {
	if _, err := validation.ToInt(value); err == nil {
		return true
	}
	if _, err := validation.ToUint(value); err == nil {
		return true
	}
	if f, err := validation.ToFloat(value); err == nil {
		return !math.IsInf(f, 0) && f == math.Trunc(f)
	}
	return false
}

// isNumericValue reports whether value is any numeric type.
func isNumericValue(value any) bool {
	if _, err := validation.ToInt(value); err == nil {
		return true
	}
	if _, err := validation.ToUint(value); err == nil {
		return true
	}
	_, err := validation.ToFloat(value)
	return err == nil
}

// isWholeJSONNumber reports whether a json.Number represents an integer value.
// A textual integer ("5") is accepted via Int64; a whole-valued decimal ("5.0")
// is accepted via Float64 so that it counts as an integer just like a
// float64-decoded 5.0 does.
func isWholeJSONNumber(n json.Number) bool {
	if _, err := n.Int64(); err == nil {
		return true
	}
	if f, err := n.Float64(); err == nil {
		return !math.IsInf(f, 0) && f == math.Trunc(f)
	}
	return false
}
