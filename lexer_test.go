package ungrammar

import (
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
		toks = append(toks, t)
		if t.name == EOF {
			break
		}
	}

	wantToks := []token{
		token{NODE, "someid", location{2, 1}},
		token{COLON, ":", location{3, 1}},
		token{QMARK, "?", location{3, 3}},
		token{NODE, "anotherid", location{3, 5}},
		token{TOKEN, "sometok", location{3, 15}},
		token{LPAREN, "(", location{5, 26}},
		token{NODE, "idmore", location{5, 28}},
		token{TOKEN, "tt tt", location{5, 35}},
		token{RPAREN, ")", location{5, 43}},
		token{TOKEN, `tt'q`, location{6, 1}},
		token{TOKEN, `tt\s`, location{6, 9}},
		token{PIPE, "|", location{7, 1}},
		token{EOF, "", location{8, 0}},
	}

	if len(wantToks) != len(toks) {
		t.Fatalf("length mismatch wantToks=%v, toks=%v", len(wantToks), len(toks))
	}
	for i := 0; i < len(wantToks); i++ {
		if wantToks[i] != toks[i] {
			t.Errorf("mismatch at index %2v: got %v, want %v", i, wantToks[i], toks[i])
		}
	}
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

// TODO: test some errors
