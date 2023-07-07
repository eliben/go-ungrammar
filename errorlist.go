// go-ungrammar: ErrorList type
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package ungrammar

import "fmt"

// ErrorList represents multiple parse errors reported by the parser on a given
// source. It's loosely modeled on scanner.ErrorList in the Go standard library.
// ErrorList implements the error interface.
type ErrorList []error

func (el *ErrorList) Add(err error) {
	*el = append(*el, err)
}

func (el ErrorList) Error() string {
	if len(el) == 0 {
		return "no errors"
	} else if len(el) == 1 {
		return el[0].Error()
	} else {
		return fmt.Sprintf("%s (and %d more errors)", el[0], len(el)-1)
	}
}
