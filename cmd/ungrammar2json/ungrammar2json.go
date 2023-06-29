package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/eliben/go-ungrammar"
)

func main() {
	if len(os.Args) != 1 {
		log.Fatal("Usage: ungrammar2json < input.ungram")
	}

	stdinBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	p := ungrammar.NewParser(string(stdinBytes))
	grammar, err := p.ParseGrammar()
	if err != nil {
		log.Fatal("Error parsing ungrammar:", err)
	}

	grammarToJSON(os.Stdout, grammar)
}

func grammarToJSON(w io.Writer, grammar *ungrammar.Grammar) {
	jrules := make(map[string]any)
	for name, rule := range grammar.Rules {
		jrules[name] = ruleToObj(rule)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(jrules); err != nil {
		log.Fatal("Error encoding to JSON:", err)
	}
}

type object map[string]any

func ruleToObj(r ungrammar.Rule) object {
	switch rr := r.(type) {
	case *ungrammar.Labeled:
		return object{"label": rr.Label, "rule": ruleToObj(rr.Rule)}
	case *ungrammar.Node:
		return object{"node": rr.Name}
	case *ungrammar.Token:
		return object{"token": rr.Value}
	case *ungrammar.Rep:
		return object{"rep": ruleToObj(rr.Rule)}
	case *ungrammar.Opt:
		return object{"opt": ruleToObj(rr.Rule)}
	case *ungrammar.Seq:
		var subRules []object
		for _, sr := range rr.Rules {
			subRules = append(subRules, ruleToObj(sr))
		}
		return object{"seq": subRules}
	case *ungrammar.Alt:
		var subRules []object
		for _, sr := range rr.Rules {
			subRules = append(subRules, ruleToObj(sr))
		}
		return object{"alt": subRules}
	default:
		return nil
	}
}
