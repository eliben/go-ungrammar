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

	p.tok = p.lex.nextToken()
	p.nextTok = p.lex.nextToken()
	return p
}

func (p *parser) parseGrammar() (*Grammar, error) {
	rules := make(map[string]Rule)
	locs := make(map[string]location)
	for !p.eof() {
		name, location, rule := p.parseNamedRule()
		if _, found := rules[name]; found {
			p.emitError(location, fmt.Sprintf("duplicate rule name %v", name))
		}
		rules[name] = rule
		locs[name] = location
	}

	grammar := &Grammar{
		Rules:   rules,
		NameLoc: locs,
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

// parseNamedRule parses a top-level named rule: Node '=' <rule>, and returns
// its name, the location of the name and the rule itself. It returns an empty
// name and rule if the parser doesn't currently point to a rule.
func (p *parser) parseNamedRule() (string, location, Rule) {
	if p.tok.name == NODE {
		tok := p.tok
		p.advance()
		if p.tok.name == EQ {
			p.advance()
			rule := p.parseAlt()
			return tok.value, tok.loc, rule
		}
	}
	p.synchronize()
	return "", location{}, nil
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
	sr := p.parseSingleRule()
	if sr == nil {
		p.emitError(p.tok.loc, fmt.Sprintf("expected rule, got %v", p.tok.value))
		return nil
	}
	seq := []Rule{sr}

	for {
		sr = p.parseSingleRule()
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

// parseSingleRule parses a single rule atom that's potentially followed by
// a '?' or '*' quantifier. It can return nil if there are no more single
// rules to parse.
//
// The Ungrammar grammr contains an ambiguity, since named rules are not
// terminated explicitly, consider:
//
//	Foo = Bar Baz
//	Bob = Rob
//
// After "Foo =" we parse a sequence of Bar, Baz, but then we see Bob, which
// shouldn't be in the sequence, but rather start a new named rule. When we
// parse a single rule, we look ahead for a '=' and bail if it's found, leaving
// "Bob =" to a higher-level parser. In that case, nil is returned.
func (p *parser) parseSingleRule() Rule {
	atom := p.parseSingleRuleAtom()
	if atom == nil {
		return nil
	}
	if p.tok.name == QMARK {
		p.advance()
		return &Opt{atom}
	} else if p.tok.name == STAR {
		p.advance()
		return &Rep{atom}
	}
	return atom
}

// parseSingleRuleAtom parser a single rule atom - either a node, token, a
// labeled rule, or a rule in parentheses. See the comment on parseSingleRule
// for the grammar ambiguity this has to handle.
func (p *parser) parseSingleRuleAtom() Rule {
	switch p.tok.name {
	case NODE:
		// Lookahead to see if this is actually the beginning of the next top-level
		// rule definition, and bail if yes.
		if p.nextTok.name == EQ {
			return nil
		} else if p.nextTok.name == COLON {
			labelTok := p.advance()
			// This is a labeled rule and the label is now in labelTok.
			// Skip the colon.
			p.advance()
			r := p.parseSingleRule()
			if r == nil {
				p.emitError(p.tok.loc, fmt.Sprintf("expected rule after label, got %v", p.tok.value))
				p.synchronize()
			}
			return &Labeled{
				Label:    labelTok.value,
				Rule:     r,
				labelLoc: labelTok.loc,
			}
		} else {
			tok := p.tok
			p.advance()
			return &Node{
				Name:    tok.value,
				nameLoc: tok.loc,
			}
		}
	case TOKEN:
		tok := p.tok
		p.advance()
		return &Token{
			Value:    tok.value,
			valueLoc: tok.loc,
		}
	case LPAREN:
		// Consume '(' and parse the full rule
		p.advance()
		r := p.parseAlt()

		// Expect closing ')', but return the rule anyway if we don't find it.
		if p.tok.name != RPAREN {
			p.emitError(p.tok.loc, fmt.Sprintf("expected ')', got %v", p.tok.value))
			p.synchronize()
			return r
		}

		// Consume ')'
		p.advance()
		return r
	}
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
