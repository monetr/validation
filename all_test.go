// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/monetr/validation"
	"github.com/stretchr/testify/assert"
)

func TestAllOf(t *testing.T) {
	t.Run("passes when every rule passes", func(t *testing.T) {
		rule := validation.AllOf(
			validation.Required,
			validation.Length(3, 10),
		)
		assert.NoError(t, rule.Validate("hello"), "a present value within the length bounds should satisfy the whole set")
	})

	t.Run("fails with the first failing rule's error", func(t *testing.T) {
		// Length is the second rule and it is the one that should trip here,
		// Required is happy because the value is present.
		rule := validation.AllOf(
			validation.Required,
			validation.Length(3, 10),
		)
		err := rule.Validate("hi")
		if assert.Error(t, err, "a value that violates one rule in the set must fail the whole set") {
			assert.Equal(t, "the length must be between 3 and 10", err.Error())
		}
	})

	t.Run("short circuits on the first failure", func(t *testing.T) {
		// Required comes first and the value is empty, so we should never even
		// reach the Length rule. If we did the message would be the length one.
		rule := validation.AllOf(
			validation.Required,
			validation.Length(3, 10),
		)
		err := rule.Validate("")
		if assert.Error(t, err, "an empty value should be rejected by Required before Length is ever consulted") {
			assert.Equal(t, "cannot be blank", err.Error())
		}
	})

	t.Run("an empty set passes anything", func(t *testing.T) {
		// AllOf with no rules is vacuously true. Not super useful but it
		// shouldnt blow up.
		assert.NoError(t, validation.AllOf().Validate("whatever"))
		assert.NoError(t, validation.AllOf().Validate(nil))
	})

	t.Run("honors Skip", func(t *testing.T) {
		// Once Skip is hit the rest of the set is ignored, so the Length rule
		// that would otherwise fail never runs.
		rule := validation.AllOf(
			validation.Skip,
			validation.Length(3, 10),
		)
		assert.NoError(t, rule.Validate("x"), "Skip should short circuit the set just like it does a normal rule list")
	})
}

// TestAllOf_AsOneOfBranch is the whole reason AllOf exists: letting a single
// field say it must match one SET of rules OR a different set. Here a key must
// either be null, OR it must be a present string between 10 and 64 characters
// that looks like a key. OneOf on its own cant express the multi-rule branch
// because it treats each alternative as a single Rule.
func TestAllOf_AsOneOfBranch(t *testing.T) {
	keyPattern := regexp.MustCompile(`^[a-z0-9_]+$`)
	rule := validation.OneOf(
		// Branch 1 -- the value is simply absent. A lone rule needs no AllOf.
		validation.Nil,
		// Branch 2 -- the value is present and satisfies all three rules.
		validation.AllOf(
			validation.Required,
			validation.Length(10, 64),
			validation.Match(keyPattern),
		),
	)

	t.Run("a nil value takes the first branch", func(t *testing.T) {
		var key *string
		assert.NoError(t, rule.Validate(key), "a missing key should be allowed by the null branch")
	})

	t.Run("a valid key takes the second branch", func(t *testing.T) {
		assert.NoError(t, rule.Validate("my_secret_key"), "a well formed key should satisfy every rule in the second branch")
	})

	t.Run("a key present with a nil value still takes the null branch", func(t *testing.T) {
		// This is the case that tripped people up: the key IS specified in the
		// map, its value just happens to be nil. Map runs the field rules
		// against that nil value, and because the value is null it should match
		// the first branch rather than being treated as "missing" or failing
		// the rules branch. A present key with an explicit null is still null.
		schema := validation.Map(
			validation.Key("key", rule),
		)
		input := map[string]any{
			"key": nil,
		}
		assert.NoError(t, schema.Validate(input), "a key whose value is explicitly nil should satisfy the null branch of the union")
	})

	t.Run("a present but malformed key matches neither branch", func(t *testing.T) {
		// "short" is present (so it is not null) but it is too short and has no
		// matching shape, so both branches reject it.
		err := rule.Validate("short")
		if assert.Error(t, err, "a present value that is too short should match neither the null branch nor the rules branch") {
			// The error should be a OneOfError carrying both branches failures.
			_, ok := err.(validation.OneOfError)
			assert.True(t, ok, "%T should be a OneOfError so a client can see each shape it could have matched", err)
		}
	})
}

// TestAllOf_WithContext makes sure the context actually reaches the rules inside
// the set. If it didnt, a context-aware rule nested in an AllOf would silently
// validate against the wrong thing.
func TestAllOf_WithContext(t *testing.T) {
	type ctxKey struct{}
	// This inner rule only passes when the expected value is stashed on the
	// context, proving the context was threaded all the way down.
	contextual := validation.WithContext(func(ctx context.Context, value any) error {
		if ctx.Value(ctxKey{}) != value {
			return validation.NewError("mismatch", "must equal the context value")
		}
		return nil
	})

	rule := validation.AllOf(validation.Required, contextual)

	t.Run("passes when the context carries the expected value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ctxKey{}, "expected")
		assert.NoError(t, rule.ValidateWithContext(ctx, "expected"), "the context value should reach the nested rule")
	})

	t.Run("fails when the context does not match", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ctxKey{}, "expected")
		err := rule.ValidateWithContext(ctx, "different")
		if assert.Error(t, err, "the nested context-aware rule should still get a say") {
			assert.Equal(t, "must equal the context value", err.Error())
		}
	})
}
