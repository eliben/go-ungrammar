package ungrammar

import (
	"fmt"
	"testing"
)

const input = `
someid
: ? anotherid
// comment
( id3 ) // doc
|
`

func TestLexer(t *testing.T) {
	lex := newLexer(input)
	var toks []token

	for {
		t := lex.nextToken()
		if t.name == EOF {
			break
		}
		toks = append(toks, t)
	}

	fmt.Println(toks)
}
