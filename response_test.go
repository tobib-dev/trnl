package trnl

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWriteHeader(t *testing.T) {

	tests := []struct {
		name        string
		have        int
		expectPanic bool
	}{
		{"Status OK", StatusOK, false},
		{"Status Created", StatusCreated, false},
		{"Status Internal Server Error", StatusInternalServerError, false},
		{"Invalid code", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			res := response{
				writer: bufio.NewWriter(&buf),
				header: Header{},
				status: tt.have,
			}

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for status code: %d, but did not panic", tt.have)
					}
				}()
			}

			res.WriteHeader(tt.have)
		})
	}
}
