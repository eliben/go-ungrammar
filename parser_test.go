package ungrammar

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// Tests parsing without errors
func TestParserTable(t *testing.T) {
	var tests = []struct {
		input     string
		wantRules []string
	}{
		// Basic rules
		{`x = mynode`, []string{`x: mynode`}},
		{`x = (mynode)`, []string{`x: mynode`}},
		{`x = mynode*`, []string{`x: Rep(mynode)`}},
		{`x = mynode?`, []string{`x: Opt(mynode)`}},
		{`x = 'atok'`, []string{`x: 'atok'`}},
		{`x = lab:mynode`, []string{`x: lab:mynode`}},
		{`x = node 'tok'`, []string{`x: Seq(node, 'tok')`}},
		{`x = foo | bar`, []string{`x: Alt(foo, bar)`}},

		// Multiple alts/seqs
		{`x = a | b | c | d | e | f`, []string{`x: Alt(a, b, c, d, e, f)`}},
		{`x = a b c   d  e     f`, []string{`x: Seq(a, b, c, d, e, f)`}},

		// Precedence between Seq and Alt and using (...)
		{`x = n | t p`, []string{`x: Alt(n, Seq(t, p))`}},
		{`x = n i | t p | i b`, []string{`x: Alt(Seq(n, i), Seq(t, p), Seq(i, b))`}},
		{`x = (n | t) p`, []string{`x: Seq(Alt(n, t), p)`}},
		{`x = (n | t) p v w | y`, []string{`x: Alt(Seq(Alt(n, t), p, v, w), y)`}},
		{`x = (n | t)? p`, []string{`x: Seq(Opt(Alt(n, t)), p)`}},
		{`x = (n | t)? p *`, []string{`x: Seq(Opt(Alt(n, t)), Rep(p))`}},

		// Misc. nesting
		{`x = (lab:Path '::')? labb:Seg`, []string{`x: Seq(Opt(Seq(lab:Path, '::')), labb:Seg)`}},
		{`x = '=='? 't' (n (',' n)* ','?)? 't'`, []string{`x: Seq(Opt('=='), 't', Opt(Seq(n, Rep(Seq(',', n)), Opt(','))), 't')`}},

		// Multiple rules
		{`x = a b y = d`, []string{`x: Seq(a, b)`, `y: d`}},
		{`x = a b c
		  y = d | t
			z = 'tok'`,
			[]string{`x: Seq(a, b, c)`, `y: Alt(d, t)`, `z: 'tok'`}},
		{`x =
			  lab:Rule 'tok'

			Rule =
			    'tok'
			  | Rule '*'`,
			[]string{`x: Seq(lab:Rule, 'tok')`, `Rule: Alt('tok', Seq(Rule, '*'))`}},
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
