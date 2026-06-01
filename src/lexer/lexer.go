package lexer

import (
	"strings"
	"unicode"

	"github.com/vortex-lang/vortex/src/common"
	"github.com/vortex-lang/vortex/src/dict"
)

type Lexer struct {
	input  []rune
	pos    int
	line   int
	col    int
	file   string
	tokens []common.Token
	start  int
}

func New(input string, file string) *Lexer {
	return &Lexer{
		input:  []rune(input),
		pos:    0,
		line:   1,
		col:    1,
		file:   file,
		tokens: make([]common.Token, 0, 256),
	}
}

func (l *Lexer) posFn() common.Position {
	return common.Position{Line: l.line, Column: l.col, File: l.file}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) lexComment() string {
	if l.peek() == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
		l.advance()
		l.advance()
		start := l.pos
		for l.pos < len(l.input) && l.input[l.pos] != '\n' {
			l.advance()
		}
		return string(l.input[start:l.pos])
	}
	if l.peek() == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '*' {
		l.advance()
		l.advance()
		depth := 1
		start := l.pos
		for l.pos < len(l.input) && depth > 0 {
			if l.peek() == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '*' {
				l.advance()
				l.advance()
				depth++
			} else if l.peek() == '*' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
				l.advance()
				l.advance()
				depth--
			} else {
				l.advance()
			}
		}
		return string(l.input[start : l.pos-2])
	}
	return ""
}

func (l *Lexer) lexIdentOrKeyword() common.Token {
	start := l.pos
	startPos := l.posFn()
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}
	word := string(l.input[start:l.pos])

	if dict.IsKeyword(word) {
		return common.Token{Kind: common.TokenKeyword, Lexeme: word, Pos: startPos}
	}

	if dict.IsNoise(word) {
		return common.Token{Kind: common.TokenIdent, Lexeme: word, Pos: startPos}
	}

	return common.Token{Kind: common.TokenIdent, Lexeme: word, Pos: startPos, Literal: word}
}

func (l *Lexer) lexNumber() common.Token {
	start := l.pos
	startPos := l.posFn()
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsDigit(ch) {
			l.advance()
		} else if ch == '.' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1]) {
			l.advance()
		} else if ch == 'e' || ch == 'E' {
			l.advance()
			if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
				l.advance()
			}
		} else {
			break
		}
	}
	val := string(l.input[start:l.pos])
	return common.Token{Kind: common.TokenNumber, Lexeme: val, Pos: startPos, Literal: val}
}

func (l *Lexer) lexString() common.Token {
	startPos := l.posFn()
	quote := l.advance()
	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.advance()
		if ch == quote {
			return common.Token{Kind: common.TokenString, Lexeme: buf.String(), Pos: startPos, Literal: buf.String()}
		}
		if ch == '\\' && l.pos < len(l.input) {
			next := l.advance()
			switch next {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case '\\':
				buf.WriteByte('\\')
			case '"':
				buf.WriteByte('"')
			case '\'':
				buf.WriteByte('\'')
			default:
				buf.WriteRune(next)
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	return common.Token{Kind: common.TokenString, Lexeme: buf.String(), Pos: startPos, Literal: buf.String()}
}

func (l *Lexer) lexSymbol() common.Token {
	startPos := l.posFn()
	ch := l.advance()
	var sym common.Symbol
	switch ch {
	case ';':
		sym = common.Semi
	case ':':
		sym = common.Colon
	case ',':
		sym = common.Comma
	case '.':
		sym = common.Dot
	case '(':
		sym = common.LParen
	case ')':
		sym = common.RParen
	case '{':
		sym = common.LBrace
	case '}':
		sym = common.RBrace
	case '[':
		sym = common.LBracket
	case ']':
		sym = common.RBracket
	case '=':
		if l.peek() == '=' {
			l.advance()
			sym = common.EqEq
		} else {
			sym = common.Equals
		}
	case '!':
		if l.peek() == '=' {
			l.advance()
			sym = common.BangEq
		} else {
			sym = common.Bang
		}
	case '<':
		if l.peek() == '=' {
			l.advance()
			sym = common.LessEq
		} else {
			sym = common.Less
		}
	case '>':
		if l.peek() == '=' {
			l.advance()
			sym = common.GreaterEq
		} else {
			sym = common.Greater
		}
	case '+':
		sym = common.Plus
	case '-':
		if l.peek() == '>' {
			l.advance()
			sym = common.Arrow
		} else {
			sym = common.Minus
		}
	case '*':
		sym = common.Star
	case '/':
		sym = common.Slash
	case '%':
		sym = common.Percent
	case '&':
		sym = common.Ampersand
	case '|':
		sym = common.Pipe
	case '^':
		sym = common.Caret
	case '~':
		sym = common.Tilde
	case '@':
		sym = common.At
	case '#':
		sym = common.Hash
	default:
		return common.Token{Kind: common.TokenSymbol, Lexeme: string(ch), Pos: startPos}
	}
	return common.Token{Kind: common.TokenSymbol, Lexeme: sym.String(), Pos: startPos, Literal: sym}
}

func (l *Lexer) Lex() []common.Token {
	for l.pos < len(l.input) {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			break
		}
		ch := l.input[l.pos]
		switch {
		case unicode.IsLetter(ch) || ch == '_':
			l.tokens = append(l.tokens, l.lexIdentOrKeyword())
		case unicode.IsDigit(ch):
			l.tokens = append(l.tokens, l.lexNumber())
		case ch == '"' || ch == '\'':
			l.tokens = append(l.tokens, l.lexString())
		case ch == '/' && l.pos+1 < len(l.input) && (l.input[l.pos+1] == '/' || l.input[l.pos+1] == '*'):
			l.lexComment()
		default:
			l.tokens = append(l.tokens, l.lexSymbol())
		}
	}
	l.tokens = append(l.tokens, common.Token{Kind: common.TokenEOF, Lexeme: "", Pos: l.posFn()})
	return l.tokens
}

func FilterNoise(tokens []common.Token) []common.Token {
	result := make([]common.Token, 0, len(tokens))
	for i, tok := range tokens {
		skip := tok.Kind == common.TokenIdent && dict.IsNoise(tok.Lexeme)
		if skip && i > 0 {
			prev := tokens[i-1]
			if prev.Kind == common.TokenKeyword && (prev.Lexeme == "let" || prev.Lexeme == "mut") {
				skip = false
			}
		}
		if skip {
			continue
		}
		result = append(result, tok)
	}
	return result
}

func Lex(input string, file string) ([]common.Token, error) {
	l := New(input, file)
	tokens := l.Lex()
	tokens = FilterNoise(tokens)
	return tokens, nil
}
