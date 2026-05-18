// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "time"

type (
	// Signed is a constraint that permits any signed integer type.
	Signed interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}

	// Unsigned is a constraint that permits any unsigned integer type.
	Unsigned interface {
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
	}

	// Integer is a constraint that permits any integer type.
	Integer interface {
		Signed | Unsigned
	}

	// Float is a constraint that permits any floating-point type.
	Float interface {
		~float32 | ~float64
	}

	// Threshold is a constraint that permits the value types supported by the
	// Min and Max rules: any integer type, any floating-point type, or time.Time.
	Threshold interface {
		Integer | Float | time.Time
	}
)
