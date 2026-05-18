// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	date0 := time.Time{}
	date20000101 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	date20001201 := time.Date(2000, 12, 1, 0, 0, 0, 0, time.UTC)
	date20000601 := time.Date(2000, 6, 1, 0, 0, 0, 0, time.UTC)

	// Note: Min is now generic (Min[T Threshold]), so the threshold type is
	// checked at compile time. Cases that previously passed an unsupported
	// threshold type (e.g. a string or struct{}) are no longer expressible.
	tests := []struct {
		tag   string
		rule  Rule
		value any
		err   string
	}{
		// int cases
		{"t1.1", Min(1), 1, ""},
		{"t1.2", Min(1), 2, ""},
		{"t1.3", Min(1), -1, "must be no less than 1"},
		{"t1.4", Min(1), 0, ""},
		{"t1.5", Min(1).Exclusive(), 1, "must be greater than 1"},
		{"t1.6", Min(1), "1", "cannot convert string to int64"},
		// uint cases
		{"t2.1", Min(uint(2)), uint(2), ""},
		{"t2.2", Min(uint(2)), uint(3), ""},
		{"t2.3", Min(uint(2)), uint(1), "must be no less than 2"},
		{"t2.4", Min(uint(2)), uint(0), ""},
		{"t2.5", Min(uint(2)).Exclusive(), uint(2), "must be greater than 2"},
		{"t2.6", Min(uint(2)), "1", "cannot convert string to uint64"},
		// float cases
		{"t3.1", Min(float64(2)), float64(2), ""},
		{"t3.2", Min(float64(2)), float64(3), ""},
		{"t3.3", Min(float64(2)), float64(1), "must be no less than 2"},
		{"t3.4", Min(float64(2)), float64(0), ""},
		{"t3.5", Min(float64(2)).Exclusive(), float64(2), "must be greater than 2"},
		{"t3.6", Min(float64(2)), "1", "cannot convert string to float64"},
		// Time cases
		{"t4.1", Min(date20000601), date20000601, ""},
		{"t4.2", Min(date20000601), date20001201, ""},
		{"t4.3", Min(date20000601), date20000101, "must be no less than 2000-06-01 00:00:00 +0000 UTC"},
		{"t4.4", Min(date20000601), date0, ""},
		{"t4.5", Min(date20000601).Exclusive(), date20000601, "must be greater than 2000-06-01 00:00:00 +0000 UTC"},
		{"t4.6", Min(date20000601).Exclusive(), 1, "cannot convert int to time.Time"},
		{"t4.8", Min(date0), date20000601, ""},
		// Json number cases
		{"t5.1", Min(1), json.Number("1"), ""},
		{"t5.2", Min(1), json.Number("2"), ""},
		{"t5.3", Min(1), json.Number("-1"), "must be no less than 1"},
		// This is so fucking stupid, 0 is considered "empty?" so even though 0 is
		// less than 1, this is considered okay?
		{"t5.4", Min(float64(1)), json.Number("0"), ""},
		{"t5.5", Min(float64(1)).Exclusive(), json.Number("1"), "must be greater than 1"},
		{"t5.6", Min(float64(1)), json.Number("1"), ""},
		{"t5.7", Min(float64(1)), json.Number("2"), ""},
		{"t5.8", Min(float64(1)), json.Number("-1"), "must be no less than 1"},
		// This is so fucking stupid, 0 is considered "empty?" so even though 0 is
		// less than 1, this is considered okay?
		{"t5.9", Min(float64(1)), json.Number("0"), ""},
		{"t5.10", Min(float64(1)).Exclusive(), json.Number("1"), "must be greater than 1"},
	}

	for _, test := range tests {
		err := test.rule.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestMinError(t *testing.T) {
	r := Min(10)
	assert.Equal(t, "must be no less than 10", r.Validate(9).Error())

	r = r.Error("123")
	assert.Equal(t, "123", r.err.Message())
}

func TestMax(t *testing.T) {
	date0 := time.Time{}
	date20000101 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	date20001201 := time.Date(2000, 12, 1, 0, 0, 0, 0, time.UTC)
	date20000601 := time.Date(2000, 6, 1, 0, 0, 0, 0, time.UTC)

	// Note: Max is now generic (Max[T Threshold]); see the comment in TestMin.
	tests := []struct {
		tag   string
		rule  Rule
		value any
		err   string
	}{
		// int cases
		{"t1.1", Max(2), 2, ""},
		{"t1.2", Max(2), 1, ""},
		{"t1.3", Max(2), 3, "must be no greater than 2"},
		{"t1.4", Max(2), 0, ""},
		{"t1.5", Max(2).Exclusive(), 2, "must be less than 2"},
		{"t1.6", Max(2), "1", "cannot convert string to int64"},
		// uint cases
		{"t2.1", Max(uint(2)), uint(2), ""},
		{"t2.2", Max(uint(2)), uint(1), ""},
		{"t2.3", Max(uint(2)), uint(3), "must be no greater than 2"},
		{"t2.4", Max(uint(2)), uint(0), ""},
		{"t2.5", Max(uint(2)).Exclusive(), uint(2), "must be less than 2"},
		{"t2.6", Max(uint(2)), "1", "cannot convert string to uint64"},
		// float cases
		{"t3.1", Max(float64(2)), float64(2), ""},
		{"t3.2", Max(float64(2)), float64(1), ""},
		{"t3.3", Max(float64(2)), float64(3), "must be no greater than 2"},
		{"t3.4", Max(float64(2)), float64(0), ""},
		{"t3.5", Max(float64(2)).Exclusive(), float64(2), "must be less than 2"},
		{"t3.6", Max(float64(2)), "1", "cannot convert string to float64"},
		// Time cases
		{"t4.1", Max(date20000601), date20000601, ""},
		{"t4.2", Max(date20000601), date20000101, ""},
		{"t4.3", Max(date20000601), date20001201, "must be no greater than 2000-06-01 00:00:00 +0000 UTC"},
		{"t4.4", Max(date20000601), date0, ""},
		{"t4.5", Max(date20000601).Exclusive(), date20000601, "must be less than 2000-06-01 00:00:00 +0000 UTC"},
		{"t4.6", Max(date20000601).Exclusive(), 1, "cannot convert int to time.Time"},
		{"t5.1", Max(2), json.Number("2"), ""},
		{"t5.2", Max(2), json.Number("1"), ""},
		{"t5.3", Max(2), json.Number("3"), "must be no greater than 2"},
		// This is so fucking stupid, 0 is considered "empty?" so even though 0 is
		// less than 1, this is considered okay?
		{"t5.4", Max(2), json.Number("0"), ""},
		{"t5.5", Max(2).Exclusive(), json.Number("2"), "must be less than 2"},
	}

	for _, test := range tests {
		err := test.rule.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestMaxError(t *testing.T) {
	r := Max(10)
	assert.Equal(t, "must be no greater than 10", r.Validate(11).Error())

	r = r.Error("123")
	assert.Equal(t, "123", r.err.Message())
}

func TestThresholdRule_ErrorObject(t *testing.T) {
	r := Max(10)
	err := NewError("code", "abc")
	r = r.ErrorObject(err)

	assert.Equal(t, err, r.err)
	assert.Equal(t, err.Code(), r.err.Code())
	assert.Equal(t, err.Message(), r.err.Message())
}
