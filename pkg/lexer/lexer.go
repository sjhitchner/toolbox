package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	TokenEOF   TokenType = -1
	TokenError           = -2

	EOF rune = -1
)

type TokenType int

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "Error: " + t.Value
	}
	if len(t.Value) > 50 {
		return fmt.Sprintf("%d:%.50q...", t.Type, t.Value)
	}
	return fmt.Sprintf("%d:%q", t.Type, t.Value)
}

type StateFunc func(*Lexer) StateFunc

type Lexer struct {
	// input  *bufio.Reader
	input   string
	start   int
	pos     int
	width   int
	stateFn StateFunc
	tokens  chan Token
}

func New(input string, initialState StateFunc) *Lexer {
	l := &Lexer{
		input:   input,
		tokens:  make(chan Token, 2),
		stateFn: initialState,
	}
	return l
}

func (t *Lexer) NextToken() Token {
	for {
		select {
		case token := <-t.tokens:
			return token
		default:
			if t.stateFn != nil {
				t.stateFn = t.stateFn(t)
			}
		}
	}
}

func (t *Lexer) Emit(i TokenType) {
	if t.pos > len(t.input) {
		t.tokens <- Token{TokenError, "Reached end of input unexpectantly"}
		return
	}

	t.tokens <- Token{
		Type:  i,
		Value: strings.TrimSpace(t.input[t.start:t.pos]),
	}
	t.start = t.pos
}

func (l *Lexer) Peek() rune {
	r := l.Next()
	l.Backup()
	return r
}

func (t *Lexer) Next() rune {
	r, w := utf8.DecodeRuneInString(t.input[t.pos:])
	t.width = w
	t.pos += t.width

	if int(t.pos) >= len(t.input) {
		t.width = 0
		return EOF
	}
	return r
}

func (t *Lexer) Skip() {
	t.Next()
	t.Ignore()
}

// Ignore skips over the pending input before this point.
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// Accept consumes the next rune
// if it's from the valid set.
func (l *Lexer) Accept(valid string) bool {
	if strings.IndexRune(valid, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

// AcceptRun consumes a run of runes from the valid set.
func (l *Lexer) AcceptRun(valid string) {
	for strings.IndexRune(valid, l.Next()) >= 0 {
	}
	l.Backup()
}

// Until consumes a run of runes until the str.
func (l *Lexer) UntilRune(r rune) {
	for l.Next() != r {
	}
	l.Backup()
}

// Until consumes a run of runes until the str.
func (l *Lexer) Until(str string) {
	if strings.HasPrefix(l.input[l.pos:], str) {
		l.pos += len(str)
	}
	l.Backup()
}

func (l *Lexer) Matches(str string) bool {
	if strings.HasPrefix(l.input[l.pos:], str) {
		l.pos += len(str)
		return true
	}
	return false
}

func (t *Lexer) Errorf(format string, args ...interface{}) StateFunc {
	t.tokens <- Token{
		Type:  TokenError,
		Value: fmt.Sprintf(format, args...),
	}
	return nil
}

func IsAlphaNumeric(r rune) bool {
	return 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' || '0' <= r && r <= '9'
}
