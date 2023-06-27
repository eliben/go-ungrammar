package ungrammar

import (
	"fmt"
	"os"
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
			gotRules := grammarToStrings(g)

			sort.Strings(tt.wantRules)
			if !slicesEqual(gotRules, tt.wantRules) {
				t.Errorf("mismatch got != want:\n%v", displaySliceDiff(gotRules, tt.wantRules))
			}
		})
	}
}

// grammarToStrings takes a Grammar's string representation and splits it into
// a sorted slice of strings (one per top-level rule) suitable for testing.
func grammarToStrings(g *Grammar) []string {
	ss := strings.Split(strings.TrimRight(g.String(), "\n"), "\n")
	sort.Strings(ss)
	return ss
}

// Check that we can read/parse ungrammar.ungrammar with some basic sanity
// checking tests.
func TestUngrammarFile(t *testing.T) {
	contents, err := os.ReadFile("./ungrammar.ungrammar")
	if err != nil {
		t.Error(err)
	}

	p := newParser(string(contents))
	g, err := p.parseGrammar()
	if err != nil {
		t.Error(err)
	}

	// TODO: use grammarToStrings here for a real test... or just add it as one
	// of the table tests somehow??
	// abstract away the grammar --> string thing and use it here too,
	// instead of this!
	if len(g.Rules) != 3 {
		t.Errorf("grammar got %v rules, want 3", len(g.Rules))
	}

	ruleAlt := g.Rules["Rule"].(*Alt)
	if len(ruleAlt.Rules) != 8 {
		t.Errorf("Rule got %v rules, want 8", len(ruleAlt.Rules))
	}
}

// TODO test errors, including lexer errors

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
