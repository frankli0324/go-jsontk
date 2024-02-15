package jsontk

type TokenType int

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

var nameOf = map[TokenType]string{
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

func (t TokenType) String() string {
	return nameOf[t]
}

type Token struct {
	Type  TokenType
	Value []byte
}
