// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "strings"

var (
	// ErrHasPrefixInvalid is the error that returns when a string does not start with the required prefix.
	ErrHasPrefixInvalid = NewError("validation_has_prefix_invalid", "must start with {{.prefix}}")
	// ErrHasSuffixInvalid is the error that returns when a string does not end with the required suffix.
	ErrHasSuffixInvalid = NewError("validation_has_suffix_invalid", "must end with {{.suffix}}")
	// ErrContainsInvalid is the error that returns when a string does not contain the required substring.
	ErrContainsInvalid = NewError("validation_contains_invalid", "must contain {{.substring}}")
)

// HasPrefix returns a validation rule that checks if a string starts with the
// given prefix. An empty value is considered valid. Use the Required rule to
// make sure a value is not empty.
func HasPrefix(prefix string) StringRule {
	return NewStringRuleWithError(
		func(s string) bool { return strings.HasPrefix(s, prefix) },
		ErrHasPrefixInvalid.SetParams(map[string]any{"prefix": prefix}),
	)
}

// HasSuffix returns a validation rule that checks if a string ends with the
// given suffix. An empty value is considered valid. Use the Required rule to
// make sure a value is not empty.
func HasSuffix(suffix string) StringRule {
	return NewStringRuleWithError(
		func(s string) bool { return strings.HasSuffix(s, suffix) },
		ErrHasSuffixInvalid.SetParams(map[string]any{"suffix": suffix}),
	)
}

// Contains returns a validation rule that checks if a string contains the given
// substring. An empty value is considered valid. Use the Required rule to make
// sure a value is not empty.
func Contains(substring string) StringRule {
	return NewStringRuleWithError(
		func(s string) bool { return strings.Contains(s, substring) },
		ErrContainsInvalid.SetParams(map[string]any{"substring": substring}),
	)
}
