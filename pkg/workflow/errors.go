package workflow

import (
	"encoding/json"
	"errors"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type requestError struct {
	detail RequestErrorDetail
	cause  error
}

func (e *requestError) Unwrap() error { return e.cause }

func (e *requestError) Error() string { raw, _ := json.Marshal(e.detail); return string(raw) }
func requestErrorAt(code, path, message string) error {
	return &validation.InputError{Err: &requestError{detail: RequestErrorDetail{Code: code, Path: path, Message: message}}}
}
func requestErrorOffset(code, path, message string, offset int) error {
	err := requestErrorAt(code, path, message)
	var detail *requestError
	errors.As(err, &detail)
	detail.detail.ByteOffset = &offset
	return err
}

// RequestErrorDetails returns a detached detail without parsing Error text.
func RequestErrorDetails(err error) (RequestErrorDetail, bool) {
	var e *requestError
	if !errors.As(err, &e) {
		return RequestErrorDetail{}, false
	}
	out := e.detail
	if out.ByteOffset != nil {
		offset := *out.ByteOffset
		out.ByteOffset = &offset
	}
	return out, true
}
func artifactError(report *environment.Report, base string, offset int) error {
	if report == nil || report.Status != "invalid" {
		return nil
	}
	for _, d := range report.Diagnostics {
		if d.Severity != "error" {
			continue
		}
		path := base + d.Path
		if d.ByteOffset != nil {
			return requestErrorOffset("request_invalid", path, d.Message, offset+*d.ByteOffset)
		}
		return requestErrorAt("request_invalid", path, d.Message)
	}
	return requestErrorAt("request_invalid", base, "invalid environment artifact")
}
