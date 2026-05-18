// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubstring(t *testing.T) {
	tests := []struct {
		tag   string
		rule  Rule
		value any
		err   string
	}{
		{"prefix1", HasPrefix("foo"), "foobar", ""},
		{"prefix2", HasPrefix("foo"), "barfoo", "must start with foo"},
		{"prefix3", HasPrefix("foo"), "", ""},
		{"prefix4", HasPrefix("foo"), []byte("foobar"), ""},
		{"suffix1", HasSuffix("bar"), "foobar", ""},
		{"suffix2", HasSuffix("bar"), "barfoo", "must end with bar"},
		{"suffix3", HasSuffix("bar"), "", ""},
		{"contains1", Contains("oob"), "foobar", ""},
		{"contains2", Contains("xyz"), "foobar", "must contain xyz"},
		{"contains3", Contains("xyz"), "", ""},
	}

	for _, test := range tests {
		err := test.rule.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestSubstring_ChainedError(t *testing.T) {
	r := HasPrefix("foo").Error("bad prefix")
	assert.Equal(t, "bad prefix", r.Validate("bar").Error())
}
