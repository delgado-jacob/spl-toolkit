package rewrite

import (
	"errors"
	"fmt"
)

// InputError marks malformed request data independently of query-level results.
type InputError struct{ Err error }

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }
func IsInputError(err error) bool   { var input *InputError; return errors.As(err, &input) }
func inputError(format string, args ...any) error {
	return &InputError{Err: fmt.Errorf(format, args...)}
}

const (
	ReasonNoMatch                = "no_match"
	ReasonConditionFalse         = "condition_false"
	ReasonNoChange               = "no_change"
	ReasonConditionUnknown       = "condition_unknown"
	ReasonUnsupportedReference   = "unsupported_reference"
	ReasonDynamicReference       = "dynamic_reference"
	ReasonConflictingTargets     = "conflicting_targets"
	ReasonOverlappingEdits       = "overlapping_edits"
	ReasonBindingCollision       = "binding_collision"
	ReasonLinkedEditUnproven     = "linked_edit_unproven"
	ReasonTargetNotRenderable    = "target_not_renderable"
	ReasonPostVerificationFailed = "post_verification_failed"
)
