package ungrammar

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestParserTable(t *testing.T) {
	var tests = []struct {
		input     string
		wantRules []string
	}{
		{`testrule = node 'tok'`, []string{`testrule: Seq(node, 'tok')`}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			g, err := p.parseGrammar()
			if err != nil {
				t.Error(err)
			}

			var gotRules []string
			for k, v := range g.Rules {
				gotRules = append(gotRules, fmt.Sprintf("%s: %s", k, v))
			}
			sort.Strings(gotRules)
			sort.Strings(tt.wantRules)

			if !slicesEqual(gotRules, tt.wantRules) {
				t.Errorf("mismatch got != want:\n%v", displaySliceDiff(gotRules, tt.wantRules))
			}
		})
	}

	//p := newParser(input)
	//g, err := p.parseGrammar()
	//if err != nil {
	//t.Error(err)
	//}

	//j, _ := json.Marshal(*g)
	//fmt.Println(string(j))
}

func displaySliceDiff[T any](got []T, want []T) string {
	maxLen := 0
	for _, g := range got {
		gs := fmt.Sprintf("%v", g)
		maxLen = intMax(maxLen+1, len(gs))
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%-*v      %v\n", maxLen, "got", "want")

	for i := 0; i < intMax(len(got), len(want)); i++ {
		var sgot string
		if i < len(got) {
			sgot = fmt.Sprintf("%v", got[i])
		}

		var swant string
		if i < len(want) {
			swant = fmt.Sprintf("%v", want[i])
		}

		sign := "  "
		if swant != sgot {
			sign = "!="
		}

		fmt.Fprintf(&sb, "%-*v  %v  %v\n", maxLen, sgot, sign, swant)
	}
	return sb.String()
}
