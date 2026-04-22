package dsl

import "fmt"

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) expect(tt TokenType) (Token, error) {
	tok := p.peek()
	if tok.Type != tt {
		return tok, fmt.Errorf("line %d:%d: expected %s, got %q", tok.Line, tok.Col, tt, tok.Value)
	}
	return p.advance(), nil
}

func Parse(src string) (*File, error) {
	tokens := Tokenize(src)
	p := NewParser(tokens)
	return p.parse()
}

func (p *Parser) parse() (*File, error) {
	file := &File{}
	seen := make(map[string]int)

	for p.peek().Type != EOF {
		def, err := p.parseDefinition()
		if err != nil {
			return nil, err
		}
		if prev, ok := seen[def.Name]; ok {
			return nil, fmt.Errorf("line %d: %q already defined at line %d", def.Line, def.Name, prev)
		}
		seen[def.Name] = def.Line
		file.Definitions = append(file.Definitions, def)
	}
	return file, nil
}

func (p *Parser) parseDefinition() (*Definition, error) {
	tok, err := p.expect(DEFINE)
	if err != nil {
		return nil, err
	}

	def := &Definition{Line: tok.Line}

	nameTok, err := p.expect(IDENT)
	if err != nil {
		return nil, err
	}
	def.Name = nameTok.Value

	if p.peek().Type == WITH {
		p.advance()
		deps, err := p.parseDeps()
		if err != nil {
			return nil, err
		}
		def.Deps = deps
	}

	if p.peek().Type == LBRACE {
		p.advance()
		invocations, err := p.parseBody()
		if err != nil {
			return nil, err
		}
		def.Invocations = invocations
		if _, err := p.expect(RBRACE); err != nil {
			return nil, err
		}
	}

	return def, nil
}

func (p *Parser) parseDeps() ([]string, error) {
	var deps []string

	nameTok, err := p.expect(IDENT)
	if err != nil {
		return nil, err
	}
	deps = append(deps, nameTok.Value)

	for p.peek().Type == COMMA {
		p.advance()
		nameTok, err := p.expect(IDENT)
		if err != nil {
			return nil, err
		}
		deps = append(deps, nameTok.Value)
	}

	return deps, nil
}

func (p *Parser) parseBody() ([]Invocation, error) {
	var invocations []Invocation

	for p.peek().Type != RBRACE && p.peek().Type != EOF {
		var targets []string
		for p.peek().Type == IDENT {
			targets = append(targets, p.advance().Value)
		}
		if len(targets) == 0 {
			tok := p.advance()
			return nil, fmt.Errorf("line %d:%d: unexpected token %q in body", tok.Line, tok.Col, tok.Value)
		}
		if _, err := p.expect(SEMICOLON); err != nil {
			return nil, err
		}
		invocations = append(invocations, Invocation{Targets: targets})
	}

	return invocations, nil
}