package ungrammar

import "fmt"

// Parser parses ungrammar syntax into a Grammar. Create a new parser with
// NewParser, and then call its ParseGrammar method.
type Parser struct {
	lex *lexer

	tok     token
	nextTok token

	errs ErrorList
}

func NewParser(buf string) *Parser {
	p := &Parser{
		lex:  newLexer(buf),
		errs: nil,
	}

	p.tok = p.lex.nextToken()
	p.nextTok = p.lex.nextToken()
	return p
}

func (p *Parser) ParseGrammar() (*Grammar, error) {
	rules := make(map[string]Rule)
	locs := make(map[string]location)
	for !p.eof() {
		name, location, rule := p.parseNamedRule()
		if rule != nil {
			if _, found := rules[name]; found {
				p.emitError(location, fmt.Sprintf("duplicate rule name %v", name))
			}
			rules[name] = rule
			locs[name] = location
		}
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
func (p *Parser) advance() token {
	tok := p.tok
	if tok.name == EOF {
		return tok
	}

	// Shift the lookahead "buffer"
	p.tok = p.nextTok
	p.nextTok = p.lex.nextToken()
	return tok
}

func (p *Parser) eof() bool {
	return p.tok.name == EOF
}

// parseNamedRule parses a top-level named rule: Node '=' <rule>, and returns
// its name, the location of the name and the rule itself. It returns an empty
// name and rule if the parser doesn't currently point to a rule.
func (p *Parser) parseNamedRule() (string, location, Rule) {
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
func (p *Parser) parseAlt() Rule {
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
func (p *Parser) parseSeq() Rule {
	sr := p.parseSingleRule()
	if sr == nil {
		p.emitError(p.tok.loc, fmt.Sprintf("expected rule, got %v", p.tok.value))
		p.synchronize()
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
// The Ungrammar grammar contains an ambiguity, since named rules are not
// terminated explicitly, consider:
//
//	Foo = Bar Baz
//	Bob = Rob
//
// After "Foo =" we parse a sequence of Bar, Baz, but then we see Bob, which
// shouldn't be in the sequence, but rather start a new named rule. When we
// parse a single rule, we look ahead for a '=' and bail if it's found, leaving
// "Bob =" to a higher-level parser. In that case, nil is returned.
func (p *Parser) parseSingleRule() Rule {
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

// parseSingleRuleAtom parses a single rule atom - either a node, token, a
// labeled rule, or a rule in parentheses. See the comment on parseSingleRule
// for the grammar ambiguity this has to handle.
func (p *Parser) parseSingleRuleAtom() Rule {
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
	case ERROR:
		p.emitError(p.tok.loc, p.tok.value)
		p.synchronize()
	}
	return nil
}

// synchronize consumes tokens until it finds a safe place to restart parsing.
// It tries to find the next Node '=' where a new named rule can be defined.
func (p *Parser) synchronize() {
	for !p.eof() {
		if p.tok.name == NODE && p.nextTok.name == EQ {
			return
		}
		p.advance()
	}
}

func (p *Parser) emitError(loc location, msg string) {
	p.errs.Add(fmt.Errorf("%s: %s", loc, msg))
}
