// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchOneOf(t *testing.T) {
	withEmail := Map(
		Key("email", Required),
		Key("password", Required),
	)
	withUsername := Map(
		Key("password", Required),
		Key("username", Required),
	)

	t.Run("first schema wins", func(t *testing.T) {
		i, err := MatchOneOf(map[string]any{
			"email":    "a@b.com",
			"password": "hunter2",
		}, withEmail, withUsername)
		assert.NoError(t, err)
		assert.Equal(t, 0, i)
	})

	t.Run("second schema wins", func(t *testing.T) {
		i, err := MatchOneOf(map[string]any{
			"username": "bob",
			"password": "hunter2",
		}, withEmail, withUsername)
		assert.NoError(t, err)
		assert.Equal(t, 1, i)
	})

	t.Run("no schema matches returns OneOfError per attempt", func(t *testing.T) {
		// A map carrying both discriminators matches neither variant: each
		// rejects the key it does not list as "key not expected". This is the
		// map equivalent of a forbidden field, with no extra rule needed.
		i, err := MatchOneOf(map[string]any{
			"email":    "a@b.com",
			"username": "bob",
			"password": "hunter2",
		}, withEmail, withUsername)
		assert.Equal(t, -1, i)

		oe, ok := err.(OneOfError)
		require.True(t, ok, "expected OneOfError, got %T", err)
		require.Len(t, oe, 2)
		assert.Equal(t, "key not expected", oe[0].(Errors)["username"].Error())
		assert.Equal(t, "key not expected", oe[1].(Errors)["email"].Error())
	})

	t.Run("internal error from a schema is surfaced directly", func(t *testing.T) {
		i, err := MatchOneOf("not a map", withEmail)
		assert.Equal(t, -1, i)
		_, isInternal := err.(InternalError)
		assert.True(t, isInternal, "got %T", err)
		_, isOneOf := err.(OneOfError)
		assert.False(t, isOneOf)
	})
}

func TestOneOfRule(t *testing.T) {
	// Two overlapping integer sets used as trivial scalar "schemas".
	low := In(1, 2)
	high := In(2, 3)

	t.Run("anyOf passes when one matches", func(t *testing.T) {
		assert.NoError(t, Validate(1, OneOf(low, high)))
	})

	t.Run("anyOf fails with OneOfError when none match", func(t *testing.T) {
		err := Validate(5, OneOf(low, high))
		oe, ok := err.(OneOfError)
		require.True(t, ok, "got %T", err)
		assert.Len(t, oe, 2)
	})

	t.Run("strict passes when exactly one matches", func(t *testing.T) {
		assert.NoError(t, Validate(1, OneOf(low, high).Strict()))
	})

	t.Run("strict fails as ambiguous when more than one matches", func(t *testing.T) {
		err := Validate(2, OneOf(low, high).Strict())
		assert.Equal(t, ErrAmbiguousMatch, err)
	})

	t.Run("strict fails with OneOfError when none match", func(t *testing.T) {
		err := Validate(5, OneOf(low, high).Strict())
		_, ok := err.(OneOfError)
		assert.True(t, ok, "got %T", err)
	})
}

func TestOneOfError_MarshalJSON(t *testing.T) {
	t.Run("single alternative still wraps in a oneOf array", func(t *testing.T) {
		oe := OneOfError{
			Errors{"name": errors.New("Name must be between 1 and 300 characters")},
		}
		raw, err := json.Marshal(oe)
		require.NoError(t, err)

		var decoded struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, []map[string]string{
			{"name": "Name must be between 1 and 300 characters"},
		}, decoded.OneOf)
	})

	t.Run("multiple alternatives", func(t *testing.T) {
		oe := OneOfError{
			Errors{"username": errors.New("must not be provided")},
			Errors{"email": errors.New("must not be provided")},
		}
		raw, err := json.Marshal(oe)
		require.NoError(t, err)

		var decoded struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, []map[string]string{
			{"username": "must not be provided"},
			{"email": "must not be provided"},
		}, decoded.OneOf)
	})

	t.Run("nested inside a parent Errors recurses", func(t *testing.T) {
		// The load-bearing case: a OneOfError sitting as a field value inside a
		// parent Errors must serialize structurally, not collapse to a string.
		parent := Errors{
			"memo": OneOfError{
				Errors{"name": errors.New("must be a valid value")},
				Errors{"derivedKind": errors.New("cannot be blank")},
			},
		}
		raw, err := json.Marshal(parent)
		require.NoError(t, err)

		var decoded map[string]struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		require.Contains(t, decoded, "memo")
		assert.Equal(t, []map[string]string{
			{"name": "must be a valid value"},
			{"derivedKind": "cannot be blank"},
		}, decoded["memo"].OneOf)
	})

	t.Run("scalar entries serialize as their message string", func(t *testing.T) {
		oe := OneOfError{
			NewError("c1", "must be a valid email address"),
			NewError("c2", "must be a valid E164 number"),
		}
		raw, err := json.Marshal(oe)
		require.NoError(t, err)

		var decoded struct {
			OneOf []string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, []string{
			"must be a valid email address",
			"must be a valid E164 number",
		}, decoded.OneOf)
	})
}

func TestOneOfError_Unwrap(t *testing.T) {
	inner := Errors{"name": ErrRequired}
	oe := OneOfError{inner, Errors{"other": ErrRequired}}

	assert.Len(t, oe.Unwrap(), 2)

	// errors.As reaches into the variants via Unwrap.
	var found Errors
	assert.True(t, errors.As(oe, &found))
	assert.Equal(t, inner, found)
}

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func TestMatchOneOfStruct(t *testing.T) {
	schemas := func(l *loginPayload) [][]*FieldRules {
		return [][]*FieldRules{
			{ // email login: username forbidden
				Field(&l.Email, Required),
				Field(&l.Password, Required),
				Field(&l.Username, Never),
			},
			{ // username login: email forbidden
				Field(&l.Email, Never),
				Field(&l.Password, Required),
				Field(&l.Username, Required),
			},
		}
	}

	t.Run("matches the email variant", func(t *testing.T) {
		l := loginPayload{Email: "a@b.com", Password: "hunter2"}
		s := schemas(&l)
		i, err := MatchOneOfStruct(context.Background(), &l, s...)
		assert.NoError(t, err)
		assert.Equal(t, 0, i)
	})

	t.Run("matches the username variant", func(t *testing.T) {
		l := loginPayload{Username: "bob", Password: "hunter2"}
		s := schemas(&l)
		i, err := MatchOneOfStruct(context.Background(), &l, s...)
		assert.NoError(t, err)
		assert.Equal(t, 1, i)
	})

	t.Run("no variant matches yields the union error", func(t *testing.T) {
		// Both discriminators supplied: each variant fails because Never rejects
		// the field it forbids.
		l := loginPayload{Email: "a@b.com", Username: "bob", Password: "hunter2"}
		s := schemas(&l)
		i, err := MatchOneOfStruct(context.Background(), &l, s...)
		assert.Equal(t, -1, i)

		oe, ok := err.(OneOfError)
		require.True(t, ok, "got %T", err)
		require.Len(t, oe, 2)
		assert.Equal(t, "must not be provided", oe[0].(Errors)["username"].Error())
		assert.Equal(t, "must not be provided", oe[1].(Errors)["email"].Error())
	})

	t.Run("union error marshals to the oneOf shape", func(t *testing.T) {
		l := loginPayload{Email: "a@b.com", Username: "bob", Password: "hunter2"}
		s := schemas(&l)
		_, err := MatchOneOfStruct(context.Background(), &l, s...)

		raw, mErr := json.Marshal(err.(OneOfError))
		require.NoError(t, mErr)

		var decoded struct {
			OneOf []map[string]string `json:"oneOf"`
		}
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, []map[string]string{
			{"username": "must not be provided"},
			{"email": "must not be provided"},
		}, decoded.OneOf)
	})
}
