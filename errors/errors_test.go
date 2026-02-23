package sublaterr

import (
	"errors"
	"fmt"
	"testing"
)

func TestTranslateError_Error(t *testing.T) {
	cases := map[string]struct {
		input  error
		output string
	}{
		"withError":    {errors.New("connection refused"), "[deepl] Translate: connection refused"},
		"withoutError": {nil, "[deepl] Translate"},
		"empty error":  {errors.New(""), "[deepl] Translate: "},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := New(ErrNetwork, "Translate", "deepl", tc.input)

			if err.Error() != tc.output {
				t.Errorf("got %q, want %q", err.Error(), tc.output)
			}
		})
	}
}

func TestTranslateError_Unwrap(t *testing.T) {
	cases := map[string]struct {
		cause error
	}{
		"withError":   {errors.New("connection refused")},
		"empty error": {errors.New("")},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			serr := New(ErrNetwork, "Translate", "deepl", tc.cause)

			if !errors.Is(serr, tc.cause) {
				t.Error("errors.Is failed to find wrapped error")
			}

			err := fmt.Errorf("wrapped: %w", serr)
			var transErr *TranslateError
			if !errors.As(err, &transErr) {
				t.Fatal("errors.As failed")
			}
			if transErr.Code != ErrNetwork {
				t.Errorf("got code %d, want %d", transErr.Code, ErrNetwork)
			}
		})
	}
}
