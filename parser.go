package ungrammar

import "fmt"

type parser struct {
	lex *lexer

	tok     token
	nextTok token

	errs ErrorList
}

func newParser(buf string) *parser {
	p := &parser{
		lex:  newLexer(buf),
		errs: nil,
	}

	p.advance()
	return p
}

func (p *parser) parseGrammar() (*Grammar, error) {
	rules := make(map[string]Rule)
	for !p.eof() {
		name, rule := p.parseNamedRule()
		if _, found := rules[name]; found {
			p.emitError(rule.Location(), fmt.Sprintf("duplicate rule name %v", name))
		}
		rules[name] = rule
	}

	grammar := &Grammar{
		Rules: rules,
	}

	if len(p.errs) > 0 {
		return grammar, p.errs
	} else {
		return grammar, nil
	}
}

// advance returns the current token and consumes it (the next call to advance
// will return the next token in the stream, etc.)
func (p *parser) advance() token {
	tok := p.tok
	if tok.name == EOF {
		return tok
	}

	// Shift the lookahead "buffer"
	p.tok = p.nextTok
	p.nextTok = p.lex.nextToken()
	return tok
}

func (p *parser) eof() bool {
	return p.tok.name == EOF
}

func (p *parser) parseNamedRule() (string, Rule) {
	if p.tok.name == NODE {
		nodeName := p.tok.value
		p.advance()
		if p.tok.name == EQ {
			p.advance()
			rule := p.parseRule()
			return nodeName, rule
		}
	}
	p.synchronize()
	return "", nil
}

func (p *parser) parseRule() Rule {
	return nil
}

func (p *parser) synchronize() {

}

func (p *parser) emitError(loc location, msg string) {
	p.errs.Add(fmt.Errorf("%d:%d: %s", loc.line, loc.column, msg))
}
