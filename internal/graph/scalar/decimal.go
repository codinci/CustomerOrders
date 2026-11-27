// Package scalar holds custom GraphQL scalar types used by the API.
package scalar

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/shopspring/decimal"
)

// ErrInvalidDecimal is returned when a value cannot be parsed as an exact decimal.
var ErrInvalidDecimal = errors.New("invalid decimal value")

// Decimal is an exact decimal amount for money. It wraps decimal.Decimal and
// implements gqlgen's MarshalGQL/UnmarshalGQL so it can be used as a scalar.
//
// On the wire it is a JSON string (e.g. "19.99"), never a float, so no value
// is rounded in transit.
type Decimal struct {
	decimal.Decimal
}

// MarshalGQL writes d as a quoted decimal string.
func (d Decimal) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(d.String()))
}

// UnmarshalGQL parses a string or integer input into a Decimal.
// Floats are rejected: by the time a float reaches us its precision is already
// lost, so clients must send money as a string such as "19.99".
// It rejects anything that is not a finite decimal.
func (d *Decimal) UnmarshalGQL(v any) error {
	var s string
	switch in := v.(type) {
	case string:
		s = in
	case int:
		s = strconv.Itoa(in)
	case int64:
		s = strconv.FormatInt(in, 10)
	default:
		return fmt.Errorf("%w: unsupported type %T (send a string)", ErrInvalidDecimal, v)
	}

	parsed, err := decimal.NewFromString(s)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidDecimal, s)
	}
	*d = Decimal{parsed}
	return nil
}
