package ungrammar

type Grammar struct {
	Rules map[string]Rule
}

type Rule interface {
	Location() location
}

type Labeled struct {
	Label    string
	Rule     Rule
	labelLoc location
}

func (lbl *Labeled) Location() location {
	return lbl.labelLoc
}

type Node struct {
	Name    string
	nameLoc location
}

func (node *Node) Location() location {
	return node.nameLoc
}

type Token struct {
	Value    string
	valueLoc location
}

func (tok *Token) Location() location {
	return tok.valueLoc
}

type Seq struct {
	Rules []Rule
}

func (seq *Seq) Location() location {
	return seq.Rules[0].Location()
}

type Alt struct {
	Rules []Rule
}

func (alt *Alt) Location() location {
	return alt.Rules[0].Location()
}

type Opt struct {
	Rule Rule
}

func (opt *Opt) Location() location {
	return opt.Rule.Location()
}

type Rep struct {
	Rule Rule
}

func (rep *Rep) Location() location {
	return rep.Rule.Location()
}
