package lexer

import (
	"fmt"
	"io"
	"os"
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

// TODO should use reader and buffer
func New(input string, initialState StateFunc) *Lexer {
	l := &Lexer{
		input:   input,
		tokens:  make(chan Token, 102),
		stateFn: initialState,
	}
	return l
}

// TODO Should use reader and buffer
func NewFromReader(reader io.Reader, initialState StateFunc) (*Lexer, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	l := &Lexer{
		input:   string(buf),
		tokens:  make(chan Token, 2),
		stateFn: initialState,
	}
	return l, nil
}

// TODO Should use reader and buffer
func NewFromFile(filename string, initialState StateFunc) (*Lexer, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	l := &Lexer{
		input:   string(buf),
		tokens:  make(chan Token, 2),
		stateFn: initialState,
	}
	return l, nil
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

func (t *Lexer) NotEOF() bool {
	return t.pos < len(t.input)
}

// Ignore skips over the pending input before this point.
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Reset resets start and pos to beginning (used for testing)
func (l *Lexer) Reset() {
	l.start = 0
	l.pos = 0
}

// Backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) Backup() {
	l.pos -= l.width
}

func (l *Lexer) BackupN(n int) {
	l.pos -= (n * l.width)
}

// Return the current buffer
func (l *Lexer) Current() string {
	return l.input[l.start:l.pos]
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
	for l.NotEOF() {
		if l.Next() == r {
			l.Backup()
			return
		}
	}
}

// Until consumes a run of runes until the str.
func (l *Lexer) Until(str string) {
	for l.NotEOF() {
		l.Next()
		if strings.HasSuffix(l.input[l.start:l.pos], str) {
			l.BackupN(len(str))
			return
		}
	}
}

// Find brings you to the beginning of the string
func (l *Lexer) Find(substring string) {
	for l.NotEOF() {
		l.Skip()
		if l.Matches(substring) {
			return
		}
	}
}

// Find a specific rune and move the next position to that char
func (l *Lexer) FindRune(r rune) {
	for l.NotEOF() {
		if l.Next() == r {
			l.Ignore()
			return
		}
	}
}

/*
	fmt.Printf("FF '%s' '%s' '%s'\n", l.input[l.start:l.pos], l.input[l.start:], string(r))
	fmt.Printf("FF '%s' '%s' '%s'\n", l.input[l.start:l.pos], l.input[l.start:], string(r))
	fmt.Printf("FF '%s' '%s' '%s'\n", l.input[l.start:l.pos], l.input[l.start:], string(r))
*/

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
