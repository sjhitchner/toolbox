package lexer

import (
	"fmt"
	"testing"

	// . "github.com/sjhitchner/toolbox/pkg/testing"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type LexerSuite struct{}

var _ = Suite(&LexerSuite{})

func (s *LexerSuite) Test(c *C) {

	lex := New("The quick brown fox jumps over the lazy dog",
		func(l *Lexer) StateFunc {
			for l.NotEOF() {
			}
			l.Emit(TokenEOF)
			return nil
		})

	c.Assert(lex.Next(), Equals, 'T')
	c.Assert(lex.Peek(), Equals, 'h')
	c.Assert(lex.Next(), Equals, 'h')
	c.Assert(lex.Next(), Equals, 'e')
	c.Assert(lex.Next(), Equals, ' ')
	lex.Backup()
	c.Assert(lex.Current(), Equals, "The")
	lex.Skip()
	c.Assert(lex.Current(), Equals, "")
	c.Assert(lex.Peek(), Equals, 'q')
	lex.UntilRune(' ')
	c.Assert(lex.Current(), Equals, "quick")
	fmt.Printf("[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.start, lex.pos)
	lex.BackupN(5)
	fmt.Printf("[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.start, lex.pos)
	c.Assert(lex.Current(), Equals, "")
	fmt.Printf("[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.start, lex.pos)
	lex.Until("fox")
	fmt.Printf("[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.start, lex.pos)
}

const (
	LeftBrace    = '{'
	RightBrace   = '}'
	LeftBracket  = '['
	RightBracket = ']'
	Colon        = ':'
	Quote        = '"'
	Comma        = ','

	KeyToken TokenType = iota
	ValueToken
	StartObject
	EndObject
	StartList
	EndList
)

func FindJSON(l *Lexer) StateFunc {
	return nil
}
