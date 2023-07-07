// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.

package ungrammar

import (
	"fmt"
	"os"
	"path/filepath"
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

		// Expected parsing of ungrammar.ungrammar
		{
			readFileOrPanic(filepath.Join("testdata", "ungrammar.ungrammar")),
			[]string{
				`Grammar: Rep(Node)`,
				`Node: Seq(name:'ident', '=', Rule)`,
				`Rule: Alt('ident', 'token_ident', Rep(Rule), Seq(Rule, Rep(Seq('|', Rule))), Seq(Rule, '?'), Seq(Rule, '*'), Seq('(', Rule, ')'), Seq(label:'ident', ':', Rule))`,
			},
		},

		{
			readFileOrPanic(filepath.Join("testdata", "exprlang.ungrammar")),
			[]string{
				`AssignStmt: Seq('set', 'ident', '=', Expr)`,
				`BinExpr: Seq(lhs:Expr, op:Alt('+', '-', '*', '/', '%'), rhs:Expr)`,
				`Expr: Alt(Literal, UnaryExpr, ParenExpr, BinExpr)`,
				`Literal: Alt('int_literal', 'ident')`,
				`ParenExpr: Seq('(', Expr, ')')`,
				`Program: Rep(Stmt)`,
				`Stmt: Alt(AssignStmt, Expr)`,
				`UnaryExpr: Seq(op:Alt('+', '-'), Expr)`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := NewParser(tt.input)
			g, err := p.ParseGrammar()
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

// Check that we can read/parse the full rust.ungrammar without errors, and
// perform basic sanity checking.
func TestRustUngrammarFile(t *testing.T) {
	contents := readFileOrPanic(filepath.Join("testdata", "rust.ungrammar"))
	p := NewParser(string(contents))
	g, err := p.ParseGrammar()
	if err != nil {
		t.Error(err)
	}
	rules := grammarToStrings(g)

	// Sanity check: the expected number of rules, and the first and last rules
	// match (note that they are first/last in string-sorted order).
	if len(rules) != 143 {
		t.Errorf("grammar got %v rules, want 143", len(g.Rules))
	}

	want0 := `Abi: Seq('extern', Opt('string'))`
	if rules[0] != want0 {
		t.Errorf("rule 0 got %v, want %v", rules[0], want0)
	}
	want142 := `YieldExpr: Seq(Rep(Attr), 'yield', Opt(Expr))`
	if rules[142] != want142 {
		t.Errorf("rule 142 got %v, want %v", rules[142], want142)
	}
}

func TestLocations(t *testing.T) {
	input := `
x = foo | bar
y = a b?`

	p := NewParser(input)
	g, err := p.ParseGrammar()
	if err != nil {
		t.Error(err)
	}

	xrule := g.Rules["x"]
	xalt := xrule.(*Alt)
	yrule := g.Rules["y"]
	yseq := yrule.(*Seq)
	yseq1opt := yseq.Rules[1].(*Opt)

	var tests = []struct {
		name          string
		loc           location
		wantLocString string
	}{
		{"x name", g.NameLoc["x"], "2:1"},
		{"x rule", xrule.Location(), "2:5"},
		{"y name", g.NameLoc["y"], "3:1"},
		{"x alt 0", xalt.Rules[0].Location(), "2:5"},
		{"x alt 1", xalt.Rules[1].Location(), "2:11"},
		{"y seq 0", yseq.Rules[0].Location(), "3:5"},
		{"y seq 1", yseq.Rules[1].Location(), "3:7"},
		{"y seq 1 opt", yseq1opt.Location(), "3:7"},
		{"y seq 1 opt rule", yseq1opt.Rule.Location(), "3:7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.loc.String() != tt.wantLocString {
				t.Errorf("got %v, want %v", tt.loc.String(), tt.wantLocString)
			}
		})
	}
}

// Test error handling and parser recovery. The parser will try to make progress
// even in face of errors, returning partial results while errors persist.
func TestParseErrors(t *testing.T) {
	var tests = []struct {
		input      string
		wantRules  []string
		wantErrors []string
	}{
		// Missing a named rule
		{`foo bar`, []string{}, []string{"1:1: expected named rule, got foo"}},

		// Missing alternation content, partial tree created with error
		{`x = a | | b`, []string{`x: Alt(a, <nil>)`}, []string{"1:9: expected rule, got |"}},

		// Missing closing ')' before new rule, but both rules created
		{`x = ( a b t = foo`, []string{`t: foo`, `x: Seq(a, b)`}, []string{"1:11: expected ')', got t"}},

		// Recovery after spurious '='
		{`x = = foo`, []string{}, []string{"1:5: expected rule, got ="}},
		{`x = = foo = y`, []string{`foo: y`}, []string{"1:5: expected rule, got ="}},

		// Duplicate rule name
		{`x = a b   x = y z`, []string{`x: Seq(y, z)`}, []string{`1:11: duplicate rule name x`}},

		// Lexer errors
		{`x = a @   y = t`, []string{`x: a`, `y: t`}, []string{"1:7: unknown token starting with '@'"}},
		{`x = a b 'two   y = t`, []string{`x: Seq(a, b)`}, []string{"1:9: unterminated token literal"}},

		// Multiple errors
		{`x = a @ y = t z = ( k`, []string{`x: a`, `y: t`, `z: k`}, []string{`1:7: unknown token starting with '@'`, `1:21: expected ')', got <end of input>`}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := NewParser(tt.input)
			g, err := p.ParseGrammar()
			gotRules := grammarToStrings(g)

			sort.Strings(tt.wantRules)
			if !slicesEqual(gotRules, tt.wantRules) {
				t.Errorf("rules mismatch got != want:\n%v", displaySliceDiff(gotRules, tt.wantRules))
			}

			if err == nil {
				t.Error("expected errors, got nil")
			}
			errlist := err.(ErrorList)
			var gotErrors []string
			for _, err := range errlist {
				gotErrors = append(gotErrors, err.Error())
			}

			if !slicesEqual(gotErrors, tt.wantErrors) {
				fmt.Println(gotErrors, tt.wantErrors)
				t.Errorf("errors mismatch got != want:\n%v", displaySliceDiff(gotErrors, tt.wantErrors))
			}
		})
	}
}

// Test the message received when multiple errors are present
func TestMultipleErrorsMessage(t *testing.T) {
	// This has two errors:
	//   - encountering the first |
	//   - unterminated '('
	input := `
foo = |
bar = ( joe
x = y`

	p := NewParser(input)
	_, err := p.ParseGrammar()
	wantErr := "2:7: expected rule, got | (and 1 more errors)"
	if err.Error() != wantErr {
		t.Errorf("got %v, want %v", err.Error(), wantErr)
	}
}

// A single isolated test useful for debugging the parser.
func TestIsolated(t *testing.T) {
	input := `x = = foo = x`
	p := NewParser(input)
	g, err := p.ParseGrammar()

	if len(g.Rules) != 1 {
		t.Errorf("got %v rules, want 1", len(g.Rules))
	}
	if err == nil {
		t.Error("got no error, want error")
	}
}

func TestIsolatedErrors(t *testing.T) {
	input := `
foo = @
bar = ( joe
x = y`
	p := NewParser(input)
	g, err := p.ParseGrammar()

	gotRules := grammarToStrings(g)

	if len(gotRules) != 2 {
		t.Errorf("got %v rules, want 2", len(gotRules))
	}
	errlist := err.(ErrorList)
	var gotErrors []string
	for _, err := range errlist {
		gotErrors = append(gotErrors, err.Error())
	}
	if len(errlist) != 3 {
		t.Errorf("got %v errors, want 3", len(errlist))
	}
}

// grammarToStrings takes a Grammar's string representation and splits it into
// a sorted slice of strings (one per top-level rule) suitable for testing.
func grammarToStrings(g *Grammar) []string {
	if len(g.String()) == 0 {
		return []string{}
	}
	ss := strings.Split(strings.TrimRight(g.String(), "\n"), "\n")
	sort.Strings(ss)
	return ss
}

// readFileOrPanic reads the given file's contents and returns them as a string.
// In case of an error, it panics.
func readFileOrPanic(filename string) string {
	contents, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(contents)
}

// displaySliceDiff displays a diff between two slices in a way that's
// readable in test output.
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
