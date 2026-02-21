package sublaterr

import (
	"fmt"
	"log/slog"
)

type ErrorCode int

const (
	ErrInvalidProvider ErrorCode = iota
	ErrInvalidLanguage
	ErrInvalidFormat
	ErrInvalidRequest
	ErrInvalidResponse
	ErrEmptyResponse
	ErrProviderAPI
	ErrHTTP
	ErrNetwork
	ErrIO
	ErrSystem
)

type TranslateError struct {
	Code     ErrorCode
	Op       string // from where this error comes from
	Provider string
	Err      error // source error if there is
}

func (e *TranslateError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("[%s] %s", e.Provider, e.Op)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Provider, e.Op, e.Err)
}

func (e *TranslateError) Unwrap() error {
	return e.Err
}

func (e *TranslateError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("code", int(e.Code)),
		slog.String("op", e.Op),
		slog.String("provider", e.Provider),
		slog.Any("cause", e.Err),
	)
}

func New(code ErrorCode, op, provider string, err error) *TranslateError {
	return &TranslateError{
		Code:     code,
		Op:       op,
		Provider: provider,
		Err:      err,
	}
}
