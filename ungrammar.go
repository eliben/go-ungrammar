package ungrammar

type Grammar struct {
	Rules map[string]Rule
}

type Rule interface {
	isRule()
}

type Labeled struct {
	Label string
	Rule  Rule
}

type Node struct {
	Name string
}

type Token struct {
	Value string
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

func (_ Labeled) isRule()
func (_ Node) isRule()
func (_ Token) isRule()
func (_ Seq) isRule()
func (_ Alt) isRule()
func (_ Opt) isRule()
func (_ Rep) isRule()
