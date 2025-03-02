package calculation

import "errors"

var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrUnsupportedOp  = errors.New("unsupported operation")
)
