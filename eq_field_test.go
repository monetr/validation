// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqField(t *testing.T) {
	password := "secret"
	tests := []struct {
		tag   string
		other string
		value any
		err   string
	}{
		{"t1", "secret", "secret", ""},
		{"t2", "secret", "wrong", "must be equal to the other value"},
		{"t3", "secret", "", ""},
		{"t4", "secret", &password, ""},
	}

	for _, test := range tests {
		other := test.other
		r := EqField(&other)
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestEqField_Struct(t *testing.T) {
	s := struct {
		Password        string
		ConfirmPassword string
	}{"hunter2", "hunter2"}

	err := ValidateStruct(&s,
		Field(&s.ConfirmPassword, EqField(&s.Password)),
	)
	assert.NoError(t, err)

	s.ConfirmPassword = "different"
	err = ValidateStruct(&s,
		Field(&s.ConfirmPassword, EqField(&s.Password)),
	)
	assert.EqualError(t, err, "ConfirmPassword: must be equal to the other value.")
}

func TestEqFieldRule_Error(t *testing.T) {
	other := "a"
	r := EqField(&other).Error("mismatch")
	assert.Equal(t, "mismatch", r.Validate("b").Error())
}

func TestEqFieldRule_ErrorObject(t *testing.T) {
	other := "a"
	r := EqField(&other)
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, "abc", r.Validate("b").Error())
}
