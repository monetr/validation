// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBetween(t *testing.T) {
	date20000101 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	date20000601 := time.Date(2000, 6, 1, 0, 0, 0, 0, time.UTC)
	date20001201 := time.Date(2000, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		tag   string
		rule  Rule
		value any
		err   string
	}{
		{"t1", Between(1, 10), 5, ""},
		{"t2", Between(1, 10), 1, ""},
		{"t3", Between(1, 10), 10, ""},
		{"t4", Between(1, 10), 11, "must be between 1 and 10"},
		{"t5", Between(2, 10), 0, ""},
		{"t6", Between(1, 10), "x", "cannot convert string to int64"},
		{"t7", Between(1, 10).Exclusive(), 1, "must be between 1 and 10"},
		{"t8", Between(1, 10).Exclusive(), 10, "must be between 1 and 10"},
		{"t9", Between(1, 10).Exclusive(), 5, ""},
		{"t10", Between(1.5, 2.5), 2.0, ""},
		{"t11", Between(1.5, 2.5), 3.0, "must be between 1.5 and 2.5"},
		{"t12", Between(date20000101, date20001201), date20000601, ""},
		{"t13", Between(date20000601, date20001201), date20000101, "must be between 2000-06-01 00:00:00 +0000 UTC and 2000-12-01 00:00:00 +0000 UTC"},
	}

	for _, test := range tests {
		err := test.rule.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestBetweenRule_ErrorObject(t *testing.T) {
	r := Between(1, 10)
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, "abc", r.Validate(20).Error())
}

func TestBetweenRule_Error(t *testing.T) {
	r := Between(1, 10).Error("out of range")
	assert.Equal(t, "out of range", r.Validate(20).Error())
}
