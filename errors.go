package jsontk

import "strconv"

type Error struct {
	Pos int
	Msg string
	Ext string
}

func (e *Error) at(pos int, ext string) *Error {
	return &Error{Msg: e.Msg, Pos: pos, Ext: ext}
}

// Is reports whether e derives from the same sentinel as target, so that
// errors.Is keeps working across At: At copies Msg from the base sentinel and
// every sentinel has a distinct Msg, so matching on Msg identifies the origin.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e.Msg == t.Msg
}

func (e *Error) Error() string {
	msg := e.Msg
	if e.Pos > -1 {
		msg = msg + " at " + strconv.Itoa(e.Pos)
	}
	if e.Ext != "" {
		msg = msg + ", " + e.Ext
	}
	return msg
}

var (
	ErrPanic              = &Error{Pos: -1, Msg: "panic occurred"}
	ErrUnexpectedSep      = &Error{Pos: -1, Msg: "invalid separator"}
	ErrEarlyEOF           = &Error{Pos: -1, Msg: "early EOF"}
	ErrInterrupt          = &Error{Pos: -1, Msg: "interrupted by user"}
	ErrUnexpectedToken    = &Error{Pos: -1, Msg: "invalid TokenType"}
	ErrInvalidParentheses = &Error{Pos: -1, Msg: "invalid parentheses"}
	ErrStandardViolation  = &Error{Pos: -1, Msg: "json not compliant to RFC8259"} // for some simple validations
	ErrInvalidJsonpath    = &Error{Pos: -1, Msg: "invalid jsonpath"}
)
