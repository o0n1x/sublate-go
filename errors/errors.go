package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidProvider  = errors.New("Invalid Provider")
	ErrInvalidLanguage  = errors.New("Invalid Language")
	ErrInvalidFormat    = errors.New("Invalid Format")
	ErrInvalidRequest   = errors.New("Invalid Request")
	ErrHTTP             = errors.New("HTTP error")
	ErrEmptyResponse    = errors.New("Empty Response")
	ErrDocumentNotFound = errors.New("Document not found")
	ErrProviderAPI      = errors.New("API error")
	ErrNetwork          = errors.New("Network error")
)

func NewErr(code error, msg string) error {
	return fmt.Errorf("%w: %s", code, msg)
}
