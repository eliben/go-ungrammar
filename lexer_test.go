package ungrammar

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	const input = `
someid
: ? anotherid 'sometok'
// comment
                         ( idmore 'tt tt' ) // doc
'tt\'q' 'tt\\s'
|
`

	lex := newLexer(input)
	var toks []token

	for {
		t := lex.nextToken()
		fmt.Println(t)
		if t.name == EOF {
			break
		}
		toks = append(toks, t)
	}

	fmt.Println(toks)
}

func TestLexerEOF(t *testing.T) {
	// Test that we get as many EOF tokens at the end of the input as we ask for.
	const input = `:  `
	lex := newLexer(input)

	if tok := lex.nextToken(); tok.name != COLON {
		t.Errorf("got %v, want COLON", tok)
	}
	for i := 0; i < 10; i++ {
		if tok := lex.nextToken(); tok.name != EOF {
			t.Errorf("got %v, want EOF", tok)
		}
	}
}
