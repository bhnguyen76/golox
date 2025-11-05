package scanner

import (
	"strconv"

	"example.com/golox/lox/shared"
)

type Scanner struct {
	source string
	tokens []Token

	start   int //Go defaults value to 0
	current int //Go defaults value to 0
	line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: source,
		tokens: make([]Token, 0),
		line:   1,
	}
}

func (s *Scanner) ScanTokens() []Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, Token{
		Type:   EOF,
		Lexeme: "",
		Line:   s.line,
	})

	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() {
	c := s.advance()

	switch c {
	case '(':
		s.addToken(LEFT_PAREN, nil)
	case ')':
		s.addToken(RIGHT_PAREN, nil)
	case '{':
		s.addToken(LEFT_BRACE, nil)
	case '}':
		s.addToken(RIGHT_BRACE, nil)
	case ',':
		s.addToken(COMMA, nil)
	case '.':
		s.addToken(DOT, nil)
	case '-':
		s.addToken(MINUS, nil)
	case '+':
		s.addToken(PLUS, nil)
	case ';':
		s.addToken(SEMICOLON, nil)
	case '*':
		s.addToken(STAR, nil)
	
	case '!':
		if s.match('=') {
			s.addToken(BANG_EQUAL, nil)
		} else {
			s.addToken(BANG, nil)
		}
	case '=':
		if s.match('=') {
			s.addToken(EQUAL_EQUAL, nil)
		} else {
			s.addToken(EQUAL, nil)
		}
	case '<':
		if s.match('=') {
			s.addToken(LESS_EQUAL, nil)
		} else {
			s.addToken(LESS, nil)
		}
	case '>':
		if s.match('=') {
			s.addToken(GREATER_EQUAL, nil)
		} else {
			s.addToken(GREATER, nil)
		}

	case '/':
		if s.match('/') {
			// A comment goes until the end of the line.
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(SLASH, nil)
		}

	case ' ', '\r', '\t':
		// ignore whitespace
    case '\n':
        s.line++

	case '"':
		s.string()

	// case 'o':
	// 	if s.match('r') {
	// 		s.addToken(OR, nil)
	// 	}

	default:
		if isDigit(c) {
			s.number()
		} else if (isAlpha(c)) {
          s.identifier();
		} else {
		shared.ErrorAt(s.line, "Unexpected character.")
		}
	}
}

func (s *Scanner) advance() byte {
	char := s.source[s.current]
	s.current++
	return char
}

func (s *Scanner) addToken(t TokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, Token{
		Type:    t,
		Lexeme:  text,
		Literal: literal,
		Line:    s.line,
	})
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() {
		return false
	}
	if s.source[s.current] != expected {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return 0 // same idea as '\0' in Java
	}
	return s.source[s.current]
}

func (s *Scanner) string() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		shared.ErrorAt(s.line, "Unterminated string.")
		return
	}

	// The closing ".
	s.advance()

	// Trim the surrounding quotes.
	value := s.source[s.start+1 : s.current-1]
	s.addToken(STRING, value)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// Consume the "."
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	// Slice the lexeme text.
	lexeme := s.source[s.start:s.current]

	// Parse to float64 (Lox numbers are doubles).
	value, err := strconv.ParseFloat(lexeme, 64)
	if err != nil {
		shared.Report(s.line, "", "Invalid number literal: "+lexeme)
		return
	}

	s.addToken(NUMBER, value)
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return 0 
	}
	return s.source[s.current+1]
}

func (s *Scanner) identifier() {
	// Consume the rest of the identifier.
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	// The full identifier text.
	text := s.source[s.start:s.current]

	if t, ok := keywords[text]; ok {
		s.addToken(t, nil)
	} else {
		s.addToken(IDENTIFIER, nil) 
	}
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}


