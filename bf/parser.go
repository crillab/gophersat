package bf

import (
	"fmt"
	"go/token"
	"io"
	"text/scanner"
)

type parser struct {
	s     scanner.Scanner
	eof   bool   // Have we reached eof yet?
	token string // Last token read
}

// Parse parses the formula from the given input Reader.
// It returns the corresponding Formula.
// Formulas are written using the following operators (from lowest to highest priority) :
//
// - for a conjunction of clauses ("and"), the ";" operator
//
// - for an equivalence, the "=" operator,
//
// - for an implication, the "->" operator,
//
// - for a disjunction ("or"), the "|" operator,
//
// - for a conjunction ("and"), the "&" operator,
//
// - for a negation, the "^" unary operator.
//
// - for an exactly-one constraint, names of variables between curly braces, eg "{a, b, c}" to specify
// exactly one of the variable a, b or c must be true.
//
// Parentheses can be used to group subformulas.
// Note there are two ways to write conjunctions, one with a low priority, one with a high priority.
// The low-priority one is useful when the user wants to describe a whole formula as a set of smaller formulas
// that must all be true.
func Parse(r io.Reader) (Formula, error) {
	var s scanner.Scanner
	s.Init(r)
	p := parser{s: s}
	p.scan()
	f, err := p.parseClause()
	if err != nil {
		return f, err
	}
	if !p.eof {
		return nil, fmt.Errorf("expected EOF, found %q at %v", p.token, p.s.Pos())
	}
	return f, nil
}

func isOperator(token string) bool {
	return token == "=" || token == "->" || token == "|" || token == "&" || token == ";"
}

func (p *parser) scan() {
	p.eof = p.eof || (p.s.Scan() == scanner.EOF)
	p.token = p.s.TokenText()
}

func (p *parser) parseClause() (f Formula, err error) {
	if isOperator(p.token) {
		return nil, fmt.Errorf("unexpected token %q at %s", p.token, p.s.Pos())
	}
	f, err = p.parseEquiv()
	if err != nil {
		return nil, err
	}
	if p.eof {
		return f, nil
	}
	if p.token == ";" {
		p.scan()
		if p.eof {
			return f, nil
		}
		f2, err := p.parseClause()
		if err != nil {
			return nil, err
		}
		return And(f, f2), nil
	}
	return f, nil
}

func (p *parser) parseEquiv() (f Formula, err error) {
	if p.eof {
		return nil, fmt.Errorf("At position %v, expected expression, found EOF", p.s.Pos())
	}
	if isOperator(p.token) {
		return nil, fmt.Errorf("unexpected token %q at %s", p.token, p.s.Pos())
	}
	f, err = p.parseImplies()
	if err != nil {
		return nil, err
	}
	if p.eof {
		return f, nil
	}
	if p.token == "=" {
		p.scan()
		if p.eof {
			return nil, fmt.Errorf("unexpected EOF")
		}
		f2, err := p.parseEquiv()
		if err != nil {
			return nil, err
		}
		return Eq(f, f2), nil
	}
	return f, nil
}

func (p *parser) parseImplies() (f Formula, err error) {
	f, err = p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.eof {
		return f, nil
	}
	if p.token == "-" {
		p.scan()
		if p.eof {
			return nil, fmt.Errorf("unexpected EOF")
		}
		if p.token != ">" {
			return nil, fmt.Errorf("invalid token %q at %v", "-"+p.token, p.s.Pos())
		}
		p.scan()
		if p.eof {
			return nil, fmt.Errorf("unexpected EOF")
		}
		f2, err := p.parseImplies()
		if err != nil {
			return nil, err
		}
		return Implies(f, f2), nil
	}
	return f, nil
}

func (p *parser) parseOr() (f Formula, err error) {
	f, err = p.parseAnd()
	if err != nil {
		return nil, err
	}
	if p.eof {
		return f, nil
	}
	if p.token == "|" {
		p.scan()
		if p.eof {
			return nil, fmt.Errorf("unexpected EOF")
		}
		f2, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		return Or(f, f2), nil
	}
	return f, nil
}

func (p *parser) parseAnd() (f Formula, err error) {
	f, err = p.parseNot()
	if err != nil {
		return nil, err
	}
	if p.eof {
		return f, nil
	}
	if p.token == "&" {
		p.scan()
		if p.eof {
			return nil, fmt.Errorf("unexpected EOF")
		}
		f2, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		return And(f, f2), nil
	}
	return f, nil
}

func (p *parser) parseNot() (f Formula, err error) {
	if isOperator(p.token) {
		return nil, fmt.Errorf("unexpected token %q at %s", p.token, p.s.Pos())
	}
	if p.token == "^" {
		p.scan()
		if p.eof {
			return nil, fmt.Errorf("unexpected EOF")
		}
		f, err = p.parseNot()
		if err != nil {
			return nil, err
		}
		return Not(f), nil
	}
	f, err = p.parseBasic()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (p *parser) parseBasic() (f Formula, err error) {
	if isOperator(p.token) || p.token == ")" {
		return nil, fmt.Errorf("unexpected token %q at %s", p.token, p.s.Pos())
	}
	if p.token == "(" {
		p.scan()
		f, err = p.parseEquiv()
		if err != nil {
			return nil, err
		}
		if p.eof {
			return nil, fmt.Errorf("expected closing parenthesis, found EOF at %s", p.s.Pos())
		}
		if p.token != ")" {
			return nil, fmt.Errorf("expected closing parenthesis, found %q at %s", p.token, p.s.Pos())
		}
		p.scan()
		return f, nil
	}
	if p.token == "{" {
		var vars []string
		for p.token != "}" {
			p.scan()
			if p.eof {
				return nil, fmt.Errorf("expected identifier, found EOF at %s", p.s.Pos())
			}
			if token.Lookup(p.token) != token.IDENT {
				return nil, fmt.Errorf("expected variable name, found %q at %s", p.token, p.s.Pos())
			}
			vars = append(vars, p.token)
			p.scan()
			if p.eof {
				return nil, fmt.Errorf("expected comma or closing brace, found EOF at %s", p.s.Pos())
			}
			if p.token != "}" && p.token != "," {
				return nil, fmt.Errorf("expected comma or closing brace, found %q at %v", p.token, p.s.Pos())
			}
		}
		p.scan()
		return Unique(vars...), nil
	}
	defer p.scan()
	return Var(p.token), nil
}
