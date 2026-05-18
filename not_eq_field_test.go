// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotEqField(t *testing.T) {
	old := "old"
	tests := []struct {
		tag   string
		other string
		value any
		err   string
	}{
		{"t1", "old", "new", ""},
		{"t2", "old", "old", "must not be equal to the other value"},
		{"t3", "old", "", ""},
		{"t4", "old", &old, "must not be equal to the other value"},
	}

	for _, test := range tests {
		other := test.other
		r := NotEqField(&other)
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestNotEqField_Struct(t *testing.T) {
	s := struct {
		OldPassword string
		NewPassword string
	}{"hunter2", "hunter2"}

	err := ValidateStruct(&s,
		Field(&s.NewPassword, NotEqField(&s.OldPassword)),
	)
	assert.EqualError(t, err, "NewPassword: must not be equal to the other value.")

	s.NewPassword = "hunter3"
	err = ValidateStruct(&s,
		Field(&s.NewPassword, NotEqField(&s.OldPassword)),
	)
	assert.NoError(t, err)
}

func TestNotEqFieldRule_Error(t *testing.T) {
	other := "a"
	r := NotEqField(&other).Error("must differ")
	assert.Equal(t, "must differ", r.Validate("a").Error())
}

func TestNotEqFieldRule_ErrorObject(t *testing.T) {
	other := "a"
	r := NotEqField(&other)
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, "abc", r.Validate("a").Error())
}
