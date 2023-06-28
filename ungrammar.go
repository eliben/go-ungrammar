package ungrammar

import (
	"fmt"
	"strings"
)

type Grammar struct {
	// Rules maps ruleName --> Rule
	Rules map[string]Rule

	// NameLoc maps ruleName --> its location in the input
	NameLoc map[string]location
}

type Rule interface {
	Location() location
	String() string
}

type Labeled struct {
	Label    string
	Rule     Rule
	labelLoc location
}

type Node struct {
	Name    string
	nameLoc location
}

type Token struct {
	Value    string
	valueLoc location
}

type Seq struct {
	Rules []Rule
}

type Alt struct {
	Rules []Rule
}

type Opt struct {
	Rule Rule
}

type Rep struct {
	Rule Rule
}

// Location methods

func (seq *Seq) Location() location {
	return seq.Rules[0].Location()
}

func (tok *Token) Location() location {
	return tok.valueLoc
}

func (node *Node) Location() location {
	return node.nameLoc
}

func (alt *Alt) Location() location {
	return alt.Rules[0].Location()
}

func (lbl *Labeled) Location() location {
	return lbl.labelLoc
}

func (opt *Opt) Location() location {
	return opt.Rule.Location()
}

func (rep *Rep) Location() location {
	return rep.Rule.Location()
}

// String methods

func (g *Grammar) String() string {
	var sb strings.Builder
	for name, rule := range g.Rules {
		fmt.Fprintf(&sb, "%s: %s\n", name, ruleString(rule))
	}
	return sb.String()
}

func (lbl *Labeled) String() string {
	return fmt.Sprintf("%s:%s", lbl.Label, ruleString(lbl.Rule))
}

func (node *Node) String() string {
	return node.Name
}

func (tok *Token) String() string {
	return fmt.Sprintf("'%s'", tok.Value)
}

func (seq *Seq) String() string {
	var parts []string
	for _, r := range seq.Rules {
		parts = append(parts, ruleString(r))
	}
	return fmt.Sprintf("Seq(%v)", strings.Join(parts, ", "))
}

func (alt *Alt) String() string {
	var parts []string
	for _, r := range alt.Rules {
		parts = append(parts, ruleString(r))
	}
	return fmt.Sprintf("Alt(%v)", strings.Join(parts, ", "))
}

func (opt *Opt) String() string {
	return fmt.Sprintf("Opt(%s)", ruleString(opt.Rule))
}

func (rep *Rep) String() string {
	return fmt.Sprintf("Rep(%s)", ruleString(rep.Rule))
}

// ruleString returns a Rule's String() representation, or <nil> if r == nil.
func ruleString(r Rule) string {
	if r == nil {
		return "<nil>"
	} else {
		return r.String()
	}
}
