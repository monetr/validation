// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"
	"encoding/json"
	"strings"
)

var (
	_ error          = OneOfError{}
	_ json.Marshaler = OneOfError{}

	// ErrAmbiguousMatch is returned by a strict union (see [OneOfRule.Strict])
	// when more than one alternative schema validates. It indicates the schemas
	// are not mutually exclusive.
	ErrAmbiguousMatch = NewError(
		"validation_oneof_ambiguous",
		"must match exactly one of the allowed shapes, but matched several",
	)
)

// OneOfError is the error returned when a value matches none of the alternative
// schemas of a union ([MatchOneOf], [OneOf] or [MatchOneOfStruct]). Entries are
// in schema order; each is normally the [Errors] produced by one schema
// attempt, though an entry may be any error (for example a nested OneOfError,
// or a scalar rule's error when a schema is a single rule rather than a
// map/struct shape).
//
// It marshals to JSON as
//
//	{"oneOf": [<entry>, <entry>, ...]}
//
// so a client can see each shape the input could have matched. A union with a
// single alternative still marshals as a one-element oneOf array — the shape is
// uniform regardless of how many alternatives there are.
//
// OneOfError is returned unwrapped (it is not boxed in another error type), so
// when it sits as a value inside a parent [Errors], [Errors.MarshalJSON]
// serializes it structurally instead of collapsing it to a string.
type OneOfError []error

// Error implements [error].
func (e OneOfError) Error() string {
	if len(e) == 0 {
		return ""
	}
	parts := make([]string, len(e))
	for i, err := range e {
		parts[i] = "(" + err.Error() + ")"
	}
	return "must match one of: " + strings.Join(parts, " or ")
}

// Unwrap exposes the per-schema errors to [errors.Is] and [errors.As].
func (e OneOfError) Unwrap() []error {
	return []error(e)
}

// MarshalJSON serializes the union failure as {"oneOf": [...]}. Each entry is
// emitted via its own [json.Marshaler] when it implements one — so an [Errors]
// entry becomes a {field: message} object and a nested OneOfError keeps its
// {"oneOf": ...} shape — otherwise it falls back to the entry's message string.
func (e OneOfError) MarshalJSON() ([]byte, error) {
	variants := make([]any, len(e))
	for i, err := range e {
		switch err := err.(type) {
		case nil:
			variants[i] = nil
		case json.Marshaler:
			variants[i] = err
		default:
			variants[i] = err.Error()
		}
	}
	return json.Marshal(map[string]any{"oneOf": variants})
}

// MatchOneOf reports the index of the first schema that fully validates value,
// or (-1, OneOfError) if none do. See [MatchOneOfWithContext] for the details.
//
// It is named MatchOneOf rather than Match because [Match] is already the
// regular-expression rule.
func MatchOneOf(value any, schemas ...Rule) (int, error) {
	return MatchOneOfWithContext(context.Background(), value, schemas...)
}

// MatchOneOfWithContext validates value against each schema in order and returns
// the index of the first one that passes. Evaluation stops at the first match
// (anyOf semantics), so earlier schemas take precedence and the schemas should
// be written to be mutually exclusive. When no schema matches it returns -1 and
// a [OneOfError] holding each schema's failure, in order.
//
// A schema is any [Rule]; the common case is a [MapRule] describing one shape
// of a map. Because a MapRule reports keys it does not list as
// [ErrKeyUnexpected], a map union forbids a field simply by omitting it from
// the variants that disallow it.
//
// If a schema reports a non-validation problem — an [InternalError] such as
// "only a map can be validated" — that error is returned directly rather than
// recorded as a failed variant, so a configuration bug is never mistaken for a
// schema mismatch.
func MatchOneOfWithContext(ctx context.Context, value any, schemas ...Rule) (int, error) {
	failures := make(OneOfError, 0, len(schemas))
	for i, schema := range schemas {
		err := validateAgainst(ctx, value, schema)
		if err == nil {
			return i, nil
		}
		if isInternalError(err) {
			return -1, err
		}
		failures = append(failures, err)
	}
	return -1, failures
}

// OneOf returns a validation [Rule] that passes when the value matches at least
// one of the given schemas. It is the rule form of [MatchOneOf] for when you
// only need pass/fail (and want to compose or nest a union inside another rule
// chain). Use [MatchOneOf] when you need to know which schema matched in order
// to act on the input. On failure the rule's error is a [OneOfError].
//
// By default OneOf uses anyOf semantics: the first matching schema wins and
// evaluation short-circuits. Call [OneOfRule.Strict] to require that exactly
// one schema match.
func OneOf(schemas ...Rule) OneOfRule {
	return OneOfRule{schemas: schemas}
}

// OneOfRule is the rule produced by [OneOf].
type OneOfRule struct {
	schemas []Rule
	strict  bool
}

// Strict returns a copy of the rule that requires exactly one schema to match.
// If more than one matches it fails with [ErrAmbiguousMatch], which flags
// schemas that are not mutually exclusive. Unlike the default anyOf behavior,
// strict mode always evaluates every schema.
func (r OneOfRule) Strict() OneOfRule {
	r.strict = true
	return r
}

// Validate implements [Rule].
func (r OneOfRule) Validate(value any) error {
	return r.ValidateWithContext(context.Background(), value)
}

// ValidateWithContext implements [RuleWithContext].
func (r OneOfRule) ValidateWithContext(ctx context.Context, value any) error {
	if !r.strict {
		_, err := MatchOneOfWithContext(ctx, value, r.schemas...)
		return err
	}

	matched := 0
	failures := make(OneOfError, 0, len(r.schemas))
	for _, schema := range r.schemas {
		err := validateAgainst(ctx, value, schema)
		if err == nil {
			matched++
			continue
		}
		if isInternalError(err) {
			return err
		}
		failures = append(failures, err)
	}
	switch {
	case matched == 1:
		return nil
	case matched > 1:
		return ErrAmbiguousMatch
	default:
		return failures
	}
}

// MatchOneOfStruct validates structPtr against several alternative field-rule
// schemas and returns the index of the first one that fully validates (anyOf
// semantics — evaluation stops at the first match). When none match it returns
// -1 and a [OneOfError] holding each schema's [Errors], in order. The returned
// index lets the caller act on the matched shape, for example to parse or merge
// it.
//
// MatchOneOfStruct is the struct counterpart to [MatchOneOf]: a struct schema
// is a []*FieldRules (as passed to [ValidateStruct]) rather than a [Rule],
// because field rules are bound to specific struct fields. Use [Never] within a
// schema to forbid the fields a variant disallows. A non-validation error from
// a schema is surfaced directly.
func MatchOneOfStruct[T any](
	ctx context.Context,
	structPtr *T,
	schemas ...[]*FieldRules,
) (int, error) {
	failures := make(OneOfError, 0, len(schemas))
	for i, schema := range schemas {
		err := ValidateStructWithContext(ctx, structPtr, schema...)
		if err == nil {
			return i, nil
		}
		if isInternalError(err) {
			return -1, err
		}
		failures = append(failures, err)
	}
	return -1, failures
}

// validateAgainst runs a single schema against value, honoring the context when
// one is supplied.
func validateAgainst(ctx context.Context, value any, schema Rule) error {
	if ctx == nil {
		return Validate(value, schema)
	}
	return ValidateWithContext(ctx, value, schema)
}

// isInternalError reports whether err is an [InternalError] wrapping a real
// (non-validation) error.
func isInternalError(err error) bool {
	ie, ok := err.(InternalError)
	return ok && ie.InternalError() != nil
}
