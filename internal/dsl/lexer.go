package dsl

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	EOF TokenType = iota
	DEFINE
	WITH
	LBRACE
	RBRACE
	COMMA
	SEMICOLON
	IDENT
)

func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case DEFINE:
		return "define"
	case WITH:
		return "with"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case COMMA:
		return ","
	case SEMICOLON:
		return ";"
	case IDENT:
		return "IDENT"
	}
	return "?"
}

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	src  []rune
	pos  int
	line int
	col  int
}

func NewLexer(src string) *Lexer {
	return &Lexer{src: []rune(src), line: 1, col: 1}
}

func (l *Lexer) peek() (rune, bool) {
	if l.pos >= len(l.src) {
		return 0, false
	}
	return l.src[l.pos], true
}

func (l *Lexer) advance() rune {
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		ch, ok := l.peek()
		if !ok {
			break
		}
		if ch == '#' {
			for {
				c, ok := l.peek()
				if !ok || c == '\n' {
					break
				}
				l.advance()
			}
		} else if unicode.IsSpace(ch) {
			l.advance()
		} else {
			break
		}
	}
}

// isNameRune returns true for characters valid inside a symbol name or path.
func isNameRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == '/' || ch == '.'
}

// isNameStart returns true for characters that can begin a symbol name.
func isNameStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '/'
}

func (l *Lexer) readIdent() string {
	var sb strings.Builder
	for {
		ch, ok := l.peek()
		if !ok || !isNameRune(ch) {
			break
		}
		sb.WriteRune(ch)
		l.advance()
	}
	return sb.String()
}

func (l *Lexer) next() Token {
	l.skipWhitespaceAndComments()

	line, col := l.line, l.col

	ch, ok := l.peek()
	if !ok {
		return Token{Type: EOF, Line: line, Col: col}
	}

	switch ch {
	case '{':
		l.advance()
		return Token{Type: LBRACE, Value: "{", Line: line, Col: col}
	case '}':
		l.advance()
		return Token{Type: RBRACE, Value: "}", Line: line, Col: col}
	case ',':
		l.advance()
		return Token{Type: COMMA, Value: ",", Line: line, Col: col}
	case ';':
		l.advance()
		return Token{Type: SEMICOLON, Value: ";", Line: line, Col: col}
	}

	if isNameStart(ch) {
		ident := l.readIdent()
		switch ident {
		case "define":
			return Token{Type: DEFINE, Value: ident, Line: line, Col: col}
		case "with":
			return Token{Type: WITH, Value: ident, Line: line, Col: col}
		}
		return Token{Type: IDENT, Value: ident, Line: line, Col: col}
	}

	l.advance()
	return Token{Type: EOF, Line: line, Col: col}
}

func Tokenize(src string) []Token {
	l := NewLexer(src)
	var tokens []Token
	for {
		tok := l.next()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}