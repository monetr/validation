// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

// This file reproduces monetr's server/datasources/table.FieldRef — a reference
// to a column in an imported document that is EITHER a named field OR a derived
// (computed) field, never both and never neither — to show the union API end to
// end against realistic input, in both its struct and map forms.

// importColumns stands in for the columns monetr pulls from the upload context
// via getColumns(ctx).
var importColumns = []string{"date", "amount", "description"}

const derivedKindRowNumber = "rowNumber"

// fieldRef mirrors table.FieldRef.
type fieldRef struct {
	Name        string `json:"name,omitempty"`
	DerivedKind string `json:"derivedKind,omitempty"`
}

// Validate is the polished, library-native version of FieldRef.Validate. It
// returns the index of the matched variant so a caller could branch on which
// shape was supplied.
//
// monetr writes this today as:
//
//	return validators.OneOfStruct(ctx, s,
//	    []*validation.FieldRules{
//	        validation.Field(&s.Name, validators.In(getColumns(ctx)...), validators.PrintableUnicode, validation.Required),
//	        validation.Field(&s.DerivedKind, validation.Empty),
//	    },
//	    []*validation.FieldRules{
//	        validation.Field(&s.Name, validation.Empty),
//	        validation.Field(&s.DerivedKind, validators.In(DerivedKindRowNumber), validation.Required),
//	    },
//	)
//
// The differences: it returns (int, error) so the matched variant is known, the
// forbidden field uses Never (clearer "must not be provided" message and proper
// pointer semantics) instead of Empty, and the OneOfError it returns is already
// JSON-ready — no MarshalErrorTree walker required.
func (s *fieldRef) Validate(ctx context.Context) (int, error) {
	return validation.MatchOneOfStruct(
		ctx,
		s,
		// Variant 0 — a named column: Name is required and must be a known
		// column; DerivedKind is forbidden.
		[]*validation.FieldRules{
			validation.Field(&s.Name,
				validation.Required,
				validation.In(importColumns...),
				is.PrintableUnicode,
			),
			validation.Field(&s.DerivedKind, validation.Never),
		},
		// Variant 1 — a derived field: DerivedKind is required and must be a
		// known kind; Name is forbidden.
		[]*validation.FieldRules{
			validation.Field(&s.Name, validation.Never),
			validation.Field(&s.DerivedKind,
				validation.Required,
				validation.In(derivedKindRowNumber),
			),
		},
	)
}

func TestUnionExample_FieldRefStruct(t *testing.T) {
	ctx := context.Background()
	inputs := []fieldRef{
		{Name: "amount"},                           // valid: a named column
		{DerivedKind: "rowNumber"},                 // valid: a derived field
		{Name: "amount", DerivedKind: "rowNumber"}, // invalid: both supplied
		{},                     // invalid: neither supplied
		{Name: "not_a_column"}, // invalid: unknown column
	}

	fmt.Println("=== struct form (MatchOneOfStruct + Never) ===")
	for _, in := range inputs {
		ref := in
		raw, _ := json.Marshal(in)
		i, err := ref.Validate(ctx)
		if err == nil {
			fmt.Printf("input %-46s => matched variant %d\n", raw, i)
			continue
		}
		out, _ := json.Marshal(err)
		fmt.Printf("input %-46s => %s\n", raw, out)
	}
}

func TestUnionExample_RequestMap(t *testing.T) {
	// The same union as a map of dynamic request data, mirroring how monetr's
	// controller.parse[T] decodes a JSON body into map[string]any and validates
	// it. In the map form a variant forbids a field simply by NOT listing it:
	// MapRule reports the unexpected key on its own, so Never is not involved.
	namedColumn := validation.Map(
		validation.Key("name", validation.Required, validation.In(importColumns...), is.PrintableUnicode),
		// "derivedKind" intentionally omitted -> forbidden in this variant.
	)
	derivedField := validation.Map(
		// "name" intentionally omitted -> forbidden in this variant.
		validation.Key("derivedKind", validation.Required, validation.In(derivedKindRowNumber)),
	)

	inputs := []map[string]any{
		{"name": "amount"},
		{"derivedKind": "rowNumber"},
		{"name": "amount", "derivedKind": "rowNumber"},
		{"name": "not_a_column"},
	}

	fmt.Println("\n=== map form (MatchOneOf, forbid by omitting the key) ===")
	for _, in := range inputs {
		raw, _ := json.Marshal(in)
		i, err := validation.MatchOneOf(in, namedColumn, derivedField)
		if err == nil {
			fmt.Printf("input %-50s => matched variant %d\n", raw, i)
			continue
		}
		// This is the entire controller response body now: because OneOfError
		// marshals itself, "problems" is just the error — no MarshalErrorTree.
		response, _ := json.Marshal(map[string]any{
			"error":    "Invalid request",
			"problems": err,
		})
		fmt.Printf("input %-50s => %s\n", raw, response)
	}
}
