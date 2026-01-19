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

func (j *Token) UnquoteBytes() ([]byte, bool) {
	return unquoteBytes(j.Value)
}

func (j *Token) UnsafeUnquote() (string, bool) {
	return unquote(j.Value)
}

func (j *Token) String() string {
	s, _ := unquoteBytes(j.Value)
	return string(s)
}

func (j *Token) UnsafeString() string {
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
