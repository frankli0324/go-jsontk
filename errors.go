package jsontk

import "errors"

var (
	ErrPanic              = errors.New("panic occurred")
	ErrUnexpectedSep      = errors.New("invalid separator")
	ErrEarlyEOF           = errors.New("early EOF")
	ErrUnexpectedToken    = errors.New("invalid TokenType")
	ErrInvalidParentheses = errors.New("invalid parentheses")
	ErrStandardViolation  = errors.New("json not compliant to RFC8259") // for some simple validations
)
