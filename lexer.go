package ungrammar

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

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

	NODE
	TOKEN

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

	NODE:  "NODE",
	TOKEN: "TOKEN",

	EQ:     "EQ",
	STAR:   "STAR",
	PIPE:   "PIPE",
	QMARK:  "QMARK",
	COLON:  "COLON",
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

	// Offset of the current rune in buf.
	rpos int

	// Offset of the next rune in buf.
	nextpos int

	lineNum int
	colNum  int
}

func newLexer(buf string) *lexer {
	lex := lexer{
		buf:     buf,
		r:       -1,
		rpos:    0,
		nextpos: 0,
		lineNum: 1,
		colNum:  1,
	}

	lex.advance()
	return &lex
}

func (lex *lexer) nextToken() token {
	lex.skipNontokens()

	if lex.r < 0 {
		return lex.emitToken(EOF, "")
	} else if isIdChar(lex.r) {
		return lex.scanNode()
	}

	switch lex.r {
	case '\'':
		return lex.scanToken()
	case '=':
		lex.advance()
		return lex.emitToken(EQ, "=")
	case '*':
		lex.advance()
		return lex.emitToken(STAR, "*")
	case '?':
		lex.advance()
		return lex.emitToken(QMARK, "?")
	case '(':
		lex.advance()
		return lex.emitToken(LPAREN, "(")
	case ')':
		lex.advance()
		return lex.emitToken(RPAREN, ")")
	case '|':
		lex.advance()
		return lex.emitToken(PIPE, "|")
	case ':':
		lex.advance()
		return lex.emitToken(COLON, ":")
	}

	return lex.emitToken(ERROR, fmt.Sprintf("unknown token starting with %q", lex.r))
}

// advance the lexer's internal state to point to the next rune in the
// input.
func (lex *lexer) advance() {
	if lex.nextpos < len(lex.buf) {
		lex.rpos = lex.nextpos
		r, w := rune(lex.buf[lex.nextpos]), 1

		if r >= utf8.RuneSelf {
			r, w = utf8.DecodeRuneInString(lex.buf[lex.nextpos:])
		}

		lex.nextpos += w
		lex.r = r
	} else {
		lex.rpos = len(lex.buf)
		lex.r = -1 // EOF
	}
}

func (lex *lexer) peekNext() rune {
	if lex.nextpos < len(lex.buf) {
		return rune(lex.buf[lex.nextpos])
	} else {
		return -1
	}
}

func (lex *lexer) emitToken(name tokenName, value string) token {
	return token{
		name:  name,
		value: value,
		loc:   location{}, // TODO fix
	}
}

func (lex *lexer) skipNontokens() {
	for {
		switch lex.r {
		case ' ', '\t', '\r':
			lex.advance()
		case '\n':
			lex.lineNum++
			lex.advance()
		case '/':
			if lex.peekNext() == '/' {
				lex.skipLineComment()
			}
		default:
			return
		}
	}
}

func (lex *lexer) skipLineComment() {
	for lex.r != '\n' && lex.r > 0 {
		lex.advance()
	}
}

func (lex *lexer) scanNode() token {
	startpos := lex.rpos
	for isIdChar(lex.r) {
		lex.advance()
	}
	return lex.emitToken(NODE, lex.buf[startpos:lex.rpos])
}

func (lex *lexer) scanToken() token {
	lex.advance() // skip leading quote
	var tokbuf strings.Builder
	for {
		if lex.r == '\'' {
			lex.advance()
			return lex.emitToken(TOKEN, tokbuf.String())
		} else if lex.r == -1 {
			return lex.emitToken(ERROR, "unterminated token literal")
		} else if lex.r == '\\' {
			if pn := lex.peekNext(); pn == '\'' || pn == '\\' {
				tokbuf.WriteRune(pn)
				lex.advance()
			} else {
				return lex.emitToken(ERROR, "invalid escape in token literal")
			}
		} else {
			tokbuf.WriteRune(lex.r)
		}
		lex.advance()
	}
}

func isIdChar(r rune) bool {
	if r >= 256 {
		return false
	}

	const mask = 0 |
		(1<<26-1)<<'A' |
		(1<<26-1)<<'a' |
		1<<'_'

	b := byte(r)
	return (uint64(1)<<b)&(mask&(1<<64-1))|(uint64(1)<<(b-64))&(mask>>64) != 0
}
