// go-ungrammar: lexical analyzer.
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.

package ungrammar

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// token represents a Ungrammar language token - it has a name (one of the
// constants declared below), string value and a location.
//
// The term "token" is slightly overloaded in this file; in Ungrammar, a quoted
// string literal is also called a "Token" -- this is just one of the kinds of
// tokens this lexer returns.
type token struct {
	name  tokenName
	value string
	loc   location
}

type location struct {
	line   int
	column int
}

func (loc location) String() string {
	return fmt.Sprintf("%v:%v", loc.line, loc.column)
}

type tokenName int

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
	return fmt.Sprintf("token{%s, '%s', %s}", tokenNames[tok.name], tok.value, tok.loc)
}

// lexer provides lexical scanning of text into Ungrammar tokens.
//
// Create a new lexer with newLexer and then call nextToken repeatedly to get
// tokens from the stream. The lexer will return an EOF token when done.
type lexer struct {
	buf string

	// Current rune.
	r rune

	// Offset of the current rune in buf.
	rpos int

	// Offset of the next rune in buf.
	nextpos int

	// location of r
	loc location
}

// newLexer creates a new lexer for the given string.
func newLexer(buf string) *lexer {
	lex := lexer{
		buf:     buf,
		r:       -1,
		rpos:    0,
		nextpos: 0,

		// column starts at 0 since advace() always increments it before we have
		// the first rune in r
		loc: location{1, 0},
	}

	lex.advance()
	return &lex
}

// nextToken returns the next token in the input string.
func (lex *lexer) nextToken() token {
	lex.skipNontokens()

	rloc := lex.loc
	if lex.r < 0 {
		return token{EOF, "<end of input>", rloc}
	} else if isIdChar(lex.r) {
		return lex.scanNode()
	}

	switch lex.r {
	case '\'':
		return lex.scanQuoted()
	case '=':
		lex.advance()
		return token{EQ, "=", rloc}
	case '*':
		lex.advance()
		return token{STAR, "*", rloc}
	case '?':
		lex.advance()
		return token{QMARK, "?", rloc}
	case '(':
		lex.advance()
		return token{LPAREN, "(", rloc}
	case ')':
		lex.advance()
		return token{RPAREN, ")", rloc}
	case '|':
		lex.advance()
		return token{PIPE, "|", rloc}
	case ':':
		lex.advance()
		return token{COLON, ":", rloc}
	default:
		errtok := lex.emitError(fmt.Sprintf("unknown token starting with %q", lex.r), rloc)
		lex.advance()
		return errtok
	}
}

// advance the lexer's internal state to point to the next rune in the
// input. advance is responsible for maintaining the main invariant of the
// lexer: at any point after advance has been called at least once, lex.r
// is the current token the lexer is looking at; lex.rpos is its offset
// the string and lex.loc is its location. lex.nextpost is the offset of the
// next token in the input. When the end of the input is reached, lex.r
// becomes EOF.
func (lex *lexer) advance() {
	if lex.nextpos < len(lex.buf) {
		lex.rpos = lex.nextpos
		r, w := rune(lex.buf[lex.nextpos]), 1

		if r >= utf8.RuneSelf {
			r, w = utf8.DecodeRuneInString(lex.buf[lex.nextpos:])
		}

		lex.nextpos += w
		lex.r = r
		lex.loc.column += 1
	} else {
		lex.rpos = len(lex.buf)
		lex.r = -1 // EOF
	}
}

// peekNext looks at the next rune in the input, after lex.r. It only works
// correctly for rune values < 128.
func (lex *lexer) peekNext() rune {
	if lex.nextpos < len(lex.buf) {
		return rune(lex.buf[lex.nextpos])
	} else {
		return -1
	}
}

func (lex *lexer) emitError(msg string, loc location) token {
	return token{
		name:  ERROR,
		value: msg,
		loc:   loc,
	}
}

func (lex *lexer) skipNontokens() {
	for {
		switch lex.r {
		case ' ', '\t', '\r':
			lex.advance()
		case '\n':
			lex.loc.line++
			// Set column to 0 because advance() immediately increments it
			lex.loc.column = 0
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
	startloc := lex.loc
	startpos := lex.rpos
	for isIdChar(lex.r) {
		lex.advance()
	}
	return token{NODE, lex.buf[startpos:lex.rpos], startloc}
}

func (lex *lexer) scanQuoted() token {
	startloc := lex.loc
	lex.advance() // skip leading quote
	var tokbuf strings.Builder
	for {
		if lex.r == '\'' {
			lex.advance()
			return token{TOKEN, tokbuf.String(), startloc}
		} else if lex.r == -1 {
			return lex.emitError("unterminated token literal", startloc)
		} else if lex.r == '\\' {
			// Skip the backslash and write the rune following it into the buffer.
			lex.advance()
			tokbuf.WriteRune(lex.r)
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
