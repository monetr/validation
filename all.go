// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "context"

// AllOf bundles several rules into a single [Rule] that passes only when the
// value satisfies every one of them. It is the AND counterpart to [OneOf]'s OR.
//
// On its own AllOf is not very interesting, applying a list of rules to a value
// is what [Validate] and a field's rule list already do. Its reason to exist is
// to be one branch of a [OneOf]: OneOf treats each alternative as a single
// [Rule], so without AllOf there is no way to say "this branch is a whole SET
// of rules". With it you can express, on a single field, that the value must
// match one set of rules OR a different set:
//
//	validation.Field(&x.Key,
//	    validation.OneOf(
//	        validation.Nil,                                            // either it is null,
//	        validation.AllOf(IsString, validation.Length(10, 64), validation.Match(keyRe)), // OR it is X and Y and Z.
//	    ),
//	)
//
// Note the single-rule branch (Nil) does not need AllOf, a lone [Rule] is
// already a valid OneOf alternative. AllOf only earns its keep on the branches
// that are made of more than one rule.
//
// Like [Validate], AllOf evaluates its rules in order and stops at the first
// one that fails, returning that rule's error. So when a OneOf branch built
// from AllOf fails, the OneOfError records the specific sub-rule that did not
// pass rather than some merged blob, which I think is the more useful thing to
// show a client. [Skip] is honored just as it is everywhere else.
func AllOf(rules ...Rule) AllOfRule {
	return AllOfRule{rules: rules}
}

// AllOfRule is the rule produced by [AllOf].
type AllOfRule struct {
	rules []Rule
}

// Validate implements [Rule].
func (r AllOfRule) Validate(value any) error {
	return r.ValidateWithContext(context.Background(), value)
}

// ValidateWithContext implements [RuleWithContext]. It threads ctx through to
// every sub-rule so a context-aware rule (or a nested [OneOf]) inside the set
// still sees it.
func (r AllOfRule) ValidateWithContext(ctx context.Context, value any) error {
	// Delegate to the package validators so that per-rule context dispatch and
	// Skip handling behave exactly the same as they do for a normal field rule
	// list. This mirrors how When and the union helpers run their inner rules.
	if ctx == nil {
		return Validate(value, r.rules...)
	}
	return ValidateWithContext(ctx, value, r.rules...)
}
