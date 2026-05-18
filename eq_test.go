// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEq(t *testing.T) {
	var v = 1
	var v2 *int
	tests := []struct {
		tag      string
		expected interface{}
		value    interface{}
		err      string
	}{
		{"t0", 1, 0, ""},
		{"t1", 1, 1, ""},
		{"t2", 1, 2, "must be equal to 1"},
		{"t3", 1, "", ""},
		{"t4", 1, "1", "must be equal to 1"},
		{"t5", 1, &v, ""},
		{"t6", 1, v2, ""},
		{"t7", "abc", "abc", ""},
		{"t8", "abc", "def", "must be equal to abc"},
	}

	for _, test := range tests {
		r := Eq(test.expected)
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_EqRule_Error(t *testing.T) {
	r := Eq(1)
	val := 4
	assert.Equal(t, "must be equal to 1", r.Validate(&val).Error())
	r = r.Error("123")
	assert.Equal(t, "123", r.err.Message())
}

func TestEqRule_ErrorObject(t *testing.T) {
	r := Eq(1)

	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
}
