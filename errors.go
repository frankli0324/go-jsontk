package jsontk

import "errors"

var (
	ErrPanic              = errors.New("panic occurred")
	ErrUnexpectedSep      = errors.New("invalid separator")
	ErrUnexpectedToken    = errors.New("invalid TokenType")
	ErrInvalidParentheses = errors.New("invalid parentheses")
)
