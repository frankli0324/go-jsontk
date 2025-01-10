package jsontk

import (
	"bytes"
	"encoding/json"
)

type TokenType uint8

const (
	INVALID TokenType = iota
	BEGIN_OBJECT
	END_OBJECT
	BEGIN_ARRAY
	END_ARRAY
	KEY
	STRING
	NUMBER
	BOOLEAN
	NULL

	cntTokenType
)

var nameOf = [cntTokenType]string{
	INVALID:      "INVALID",
	BEGIN_OBJECT: "BEGIN_OBJECT",
	END_OBJECT:   "END_OBJECT",
	BEGIN_ARRAY:  "BEGIN_ARRAY",
	END_ARRAY:    "END_ARRAY",
	KEY:          "KEY",
	STRING:       "STRING",
	NUMBER:       "NUMBER",
	BOOLEAN:      "BOOLEAN",
	NULL:         "NULL",
}

var assuredToken = [cntTokenType]string{
	INVALID:      "<Invalid Token>",
	BEGIN_OBJECT: "{",
	END_OBJECT:   "}",
	BEGIN_ARRAY:  "[",
	END_ARRAY:    "]",
	NULL:         "null",
}

var commaAfterToken = [cntTokenType]bool{
	INVALID:      false,
	BEGIN_OBJECT: false,
	END_OBJECT:   true,
	BEGIN_ARRAY:  false,
	END_ARRAY:    true,
	KEY:          false,
	STRING:       true,
	NUMBER:       true,
	BOOLEAN:      true,
	NULL:         true,
}

func (t TokenType) String() string {
	return nameOf[t]
}

type Token struct {
	Type  TokenType
	Value []byte
}

func (t *Token) AppendTo(data []byte) []byte {
	if s := assuredToken[t.Type]; s != "" {
		return append(data, s...)
	}
	return append(data, t.Value...)
}

func (j *Token) Number() json.Number {
	return json.Number(j.Value)
}

// Unquote unquotes the underlying value as quoted string, and returns if it's successfully unquoted
func (j *Token) Unquote() (string, bool) {
	return unquote(j.Value)
}

// String behaves the same way as [Unquote] but not check for results, returns empty string on invalid results
func (j *Token) String() string {
	s, _ := unquote(j.Value)
	return s
}

func (j *Token) EqualString(s string) bool {
	return unquotedEqualStr(j.Value, s)
}

func (j *Token) Bool() bool {
	// since it's successfully tokenized, values should be always certain
	return bytes.Equal(j.Value, []byte("true"))
}
