package ungrammar

import "fmt"

type location struct {
	line   int
	column int
}

// tokenName is a type for describing tokens mnemonically.
type tokenName int

type token struct {
	name  tokenName
	value string
	loc   location
}

const (
	// Special tokens
	ERROR tokenName = iota
	EOF

	EQ
	STAR
	PIPE
	QMARK
	COLON
	LPAREN
	RPAREN
)

var tokenNames = [...]string{
	ERROR: "ERROR",
	EOF:   "EOF",

	EQ:     "EQ",
	STAR:   "STAR",
	PIPE:   "PIPE",
	QMARK:  "QMARK",
	LPAREN: "LPAREN",
	RPAREN: "RPAREN",
}

func (tok token) String() string {
	return fmt.Sprintf("token{%s, '%s', (%v, %v)}", tokenNames[tok.name], tok.value, tok.loc.line, tok.loc.column)
}

// lexer
//
// Create a new lexer with newLexer and then call nextToken repeatedly to get
// tokens from the stream. The lexer will return a token with the name EOF when
// done.
type lexer struct {
	buf string

	// Current rune.
	r rune

	// Offest of the current rune in buf.
	rpos int

	// Offset of the next rune in buf.
	nextpos int

	lineNum int
}

func newLexer(buf string) *lexer {
	lex := lexer{
		buf:     buf,
		r:       -1,
		rpos:    0,
		nextpos: 0,
		lineNum: 1,
	}

	// Prime the lexer by calling .next
	lex.next()
	return &lex
}
