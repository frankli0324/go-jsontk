package jsontk

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

func (t Token) AppendTo(data []byte) []byte {
	if s := assuredToken[t.Type]; s != "" {
		return append(data, s...)
	}
	return append(data, t.Value...)
}
