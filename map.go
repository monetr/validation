package validation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrNotMap is the error that the value being validated is not a map.
	ErrNotMap = errors.New("only a map can be validated")

	// ErrKeyWrongType is the error returned in case of an incorrect key type.
	ErrKeyWrongType = NewError("validation_key_wrong_type", "key not the correct type")

	// ErrKeyMissing is the error returned in case of a missing key.
	ErrKeyMissing = NewError("validation_key_missing", "required key is missing")

	// ErrKeyUnexpected is the error returned in case of an unexpected key.
	ErrKeyUnexpected = NewError("validation_key_unexpected", "key not expected")
)

type (
	// MapRule represents a rule set associated with a map.
	MapRule[K comparable] struct {
		keys           []*KeyRules[K]
		allowExtraKeys bool
	}

	// KeyRules represents a rule set associated with a map key.
	KeyRules[K comparable] struct {
		key      K
		optional bool
		rules    []Rule
	}
)

// Map returns a validation rule that checks the keys and values of a map.
// This rule should only be used for validating maps, or a validation error will be reported.
// Use Key() to specify map keys that need to be validated. Each Key() call specifies a single key which can
// be associated with multiple rules.
// For example,
//
//	validation.Map(
//	    validation.Key("Name", validation.Required),
//	    validation.Key("Value", validation.Required, validation.Length(5, 10)),
//	)
//
// A nil value is considered valid. Use the Required rule to make sure a map value is present.
func Map[K comparable](keys ...*KeyRules[K]) MapRule[K] {
	return MapRule[K]{keys: keys}
}

// AllowExtraKeys configures the rule to ignore extra keys.
func (r MapRule[K]) AllowExtraKeys() MapRule[K] {
	r.allowExtraKeys = true
	return r
}

// Validate checks if the given value is valid or not.
func (r MapRule[K]) Validate(m any) error {
	return r.ValidateWithContext(nil, m)
}

// ValidateWithContext checks if the given value is valid or not.
func (r MapRule[K]) ValidateWithContext(ctx context.Context, m any) error {
	value := reflect.ValueOf(m)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Map {
		// must be a map
		return NewInternalError(ErrNotMap)
	}
	if value.IsNil() {
		// treat a nil map as valid
		return nil
	}

	errs := Errors{}
	kt := value.Type().Key()

	var extraKeys map[any]bool
	if !r.allowExtraKeys {
		extraKeys = make(map[any]bool, value.Len())
		for _, k := range value.MapKeys() {
			extraKeys[k.Interface()] = true
		}
	}

	for _, kr := range r.keys {
		var err error
		if kv := reflect.ValueOf(kr.key); !kt.AssignableTo(kv.Type()) {
			err = ErrKeyWrongType
		} else if vv := value.MapIndex(kv); !vv.IsValid() {
			if !kr.optional {
				err = ErrKeyMissing
			}
		} else if ctx == nil {
			err = Validate(vv.Interface(), kr.rules...)
		} else {
			err = ValidateWithContext(ctx, vv.Interface(), kr.rules...)
		}
		if err != nil {
			if ie, ok := err.(InternalError); ok && ie.InternalError() != nil {
				return err
			}
			errs[getErrorKeyName(kr.key)] = err
		}
		if !r.allowExtraKeys {
			delete(extraKeys, kr.key)
		}
	}

	if !r.allowExtraKeys {
		for key := range extraKeys {
			errs[getErrorKeyName(key)] = ErrKeyUnexpected
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Key specifies a map key and the corresponding validation rules.
func Key[K comparable](key K, rules ...Rule) *KeyRules[K] {
	return &KeyRules[K]{
		key:   key,
		rules: rules,
	}
}

// Deprecated: Use Required instead.
//
// Optional configures the rule to ignore the key if missing.
func (r *KeyRules[K]) Optional() *KeyRules[K] {
	r.optional = true
	return r
}

// Required sets whether or not this key is required. If it is optional then you
// can pass false to this function. Not calling this function will default to
// the key being required.
func (r *KeyRules[K]) Required(required bool) *KeyRules[K] {
	r.optional = !required
	return r
}

// getErrorKeyName returns the name that should be used to represent the validation error of a map key.
func getErrorKeyName(key any) string {
	return fmt.Sprintf("%v", key)
}
