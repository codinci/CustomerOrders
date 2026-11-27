package scalar_test

import (
	"errors"
	"testing"

	"github.com/codinci/customerorders/internal/graph/scalar"
)

func TestDecimalUnmarshalGQL(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{name: "string keeps exact digits", input: "19.99", want: "19.99"},
		{name: "string keeps trailing zeros", input: "19.9900", want: "19.99"},
		{name: "int", input: 7, want: "7"},
		{name: "int64", input: int64(-3), want: "-3"},
		{name: "float64 is rejected", input: 19.99, wantErr: true},
		{name: "bool is rejected", input: true, wantErr: true},
		{name: "malformed string is rejected", input: "19.99.00", wantErr: true},
		{name: "empty string is rejected", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d scalar.Decimal
			err := d.UnmarshalGQL(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalGQL(%v) = %v, want error", tt.input, d)
				}
				if !errors.Is(err, scalar.ErrInvalidDecimal) {
					t.Fatalf("UnmarshalGQL(%v) error = %v, want ErrInvalidDecimal", tt.input, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnmarshalGQL(%v): %v", tt.input, err)
			}
			if got := d.String(); got != tt.want {
				t.Fatalf("UnmarshalGQL(%v) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}
