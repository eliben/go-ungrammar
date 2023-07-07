// go-ungrammar: basic usage example.
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package ungrammar_test

import (
	"fmt"

	"github.com/eliben/go-ungrammar"
)

func ExampleParseAndExamine() {
	input := `
Foo = Bar Baz
Baz = ( Kay Jay )* | 'id'`

	// Create an Ungrammar parser and parse input.
	p := ungrammar.NewParser(input)
	ungram, err := p.ParseGrammar()
	if err != nil {
		panic(err)
	}

	// Display the string representation of the parsed ungrammar.
	fmt.Println(ungram.Rules["Foo"].String())
	fmt.Println(ungram.Rules["Baz"].String())
	// Output:
	// Seq(Bar, Baz)
	// Alt(Rep(Seq(Kay, Jay)), 'id')
}
