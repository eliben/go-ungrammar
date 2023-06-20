package ungrammar

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	const input = `darule = node 'tok'`

	p := newParser(input)
	g, err := p.parseGrammar()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(*g)
	fmt.Println(g.Rules["darule"])

	j, _ := json.Marshal(*g)
	fmt.Println(string(j))
}
