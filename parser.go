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
			rule := p.parseAlt()
			return nodeName, rule
		}
	}
	p.synchronize()
	return "", nil
}

// parseAlt parses a top-level rule, the LHS of Node '=' <Rule>. It's
// potentially a '|'-seprated alternation of sequences.
func (p *parser) parseAlt() Rule {
	alts := []Rule{p.parseSeq()}
	for p.tok.name == PIPE {
		p.advance()
		alts = append(alts, p.parseSeq())
	}
	if len(alts) == 1 {
		return alts[0]
	} else {
		return &Alt{alts}
	}
}

// parseSeq parses a sequence of single rules.
func (p *parser) parseSeq() Rule {
	seq := []Rule{p.parseSingleRule()}

	for {
		sr := p.parseSingleRule()
		if sr == nil {
			break
		}
		seq = append(seq, sr)
	}
	if len(seq) == 1 {
		return seq[0]
	} else {
		return &Seq{seq}
	}
}

func (p *parser) parseSingleRule() Rule {
	return nil
}

// synchronize consumes tokens until it finds a safe place to restart parsing.
// It tries to find the next Node '=' where a new named rule can be defined.
func (p *parser) synchronize() {
	for !p.eof() {
		if p.tok.name == NODE && p.nextTok.name == EQ {
			return
		}
		p.advance()
	}
}

func (p *parser) emitError(loc location, msg string) {
	p.errs.Add(fmt.Errorf("%d:%d: %s", loc.line, loc.column, msg))
}
