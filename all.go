// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import "context"

// ErrAllOfInvalid is the error a custom-messaged [AllOf] surfaces when any rule
// in the set fails. By default AllOf has no error of its own (it just passes
// the first failing sub-rule's error straight through), so this only comes into
// play once you set a custom message with [AllOfRule.Error].
var ErrAllOfInvalid = NewError("validation_all_of_invalid", "must be in a valid format")

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
//
// That per-sub-rule error is the right default, but sometimes it is the wrong
// thing to leak. For a compound validation made of a pile of fiddly rules (an
// RRULE, a cron expression, some gnarly ID format) the individual sub-rule
// errors are noise to the client. [AllOfRule.Error] lets you collapse ANY
// failure within the set into a single summary message instead, so you can just
// say "yeah thats not right" and not explain which specific piece tripped.
func AllOf(rules ...Rule) AllOfRule {
	return AllOfRule{rules: rules}
}

// AllOfRule is the rule produced by [AllOf].
type AllOfRule struct {
	rules []Rule
	// err is the custom error to surface when ANY sub-rule fails. When it is nil
	// (the default) AllOf stays transparent and returns the first failing
	// sub-rule's own error instead.
	err Error
}

// Error sets a custom error message for the whole set. Normally AllOf is
// transparent: when a sub-rule fails its specific error is what bubbles up.
// Setting a message here flips that, so ANY failure within the set collapses
// into this one summary. This is meant for nuanced compound validations where
// the inner rule errors are an implementation detail the client does not care
// about, and "yeah thats not right" is a kinder thing to surface than whichever
// regex or length check happened to trip.
//
// A real misconfiguration (an [InternalError] from a sub-rule, like "only a map
// can be validated") is still surfaced as-is and NOT masked, otherwise a
// genuine bug would hide behind the friendly message.
func (r AllOfRule) Error(message string) AllOfRule {
	if r.err == nil {
		r.err = ErrAllOfInvalid
	}
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the whole error struct to surface when any sub-rule fails,
// the same way [AllOfRule.Error] sets just the message. Pass a fully built
// [Error] when you also need to control the translation code or params.
func (r AllOfRule) ErrorObject(err Error) AllOfRule {
	r.err = err
	return r
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
	var err error
	if ctx == nil {
		err = Validate(value, r.rules...)
	} else {
		err = ValidateWithContext(ctx, value, r.rules...)
	}
	if err == nil {
		return nil
	}

	// No custom error means we stay transparent and surface the specific sub-rule
	// failure, which is the original AllOf behavior.
	if r.err == nil {
		return err
	}

	// We have a custom summary error, but we do NOT want to swallow a genuine
	// misconfiguration behind it. An InternalError is a programmer or config
	// problem rather than the value being invalid, so let it through untouched
	// just like OneOf does.
	if isInternalError(err) {
		return err
	}

	return r.err
}
