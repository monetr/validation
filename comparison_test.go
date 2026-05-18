// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparison(t *testing.T) {
	tests := []struct {
		tag   string
		rule  Rule
		value any
		err   string
	}{
		// Gt
		{"gt1", Gt(10), 11, ""},
		{"gt2", Gt(10), 10, "must be greater than 10"},
		{"gt3", Gt(10), 9, "must be greater than 10"},
		{"gt4", Gt(10), 0, ""},
		// Gte
		{"gte1", Gte(10), 11, ""},
		{"gte2", Gte(10), 10, ""},
		{"gte3", Gte(10), 9, "must be no less than 10"},
		// Lt
		{"lt1", Lt(10), 9, ""},
		{"lt2", Lt(10), 10, "must be less than 10"},
		{"lt3", Lt(10), 11, "must be less than 10"},
		// Lte
		{"lte1", Lte(10), 9, ""},
		{"lte2", Lte(10), 10, ""},
		{"lte3", Lte(10), 11, "must be no greater than 10"},
	}

	for _, test := range tests {
		err := test.rule.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestComparison_ChainedError(t *testing.T) {
	r := Gt(10).Error("too small")
	assert.Equal(t, "too small", r.Validate(5).Error())
}
