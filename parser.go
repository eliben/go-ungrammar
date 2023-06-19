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
		if _, found := rules[name]; ok {
			j
		}
		rules[name] = rule
	}

	if len(p.errs) > 0 {
		return rules, p.errs
	} else {
		return rules, nil
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

func (p *parser) parseNamedRules() map[string]Rule {
}

func (p *parser) emitError(loc location, msg string) {
	p.errs.Add(fmt.Errorf("%d:%d: %s", loc.line, loc.column, msg))
}
