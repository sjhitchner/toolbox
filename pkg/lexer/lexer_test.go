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
	lex.BackupN(5)
	c.Assert(lex.Current(), Equals, "")
	lex.Until("fox")
	c.Assert(lex.Current(), Equals, "quick brown ")
	lex.Reset()
	c.Assert(lex.Accept("The"), Equals, true)
	c.Assert(lex.Accept("The"), Equals, true)
	c.Assert(lex.Accept("The"), Equals, true)
	c.Assert(lex.Accept("The"), Equals, false)
	c.Assert(lex.Current(), Equals, "The")
	lex.Next()
	lex.AcceptRun("quick")
	c.Assert(lex.Current(), Equals, "The quick")
	lex.Find("fox")
	c.Assert(lex.Current(), Equals, "fox")
	lex.Skip()
	lex.UntilRune(' ')
	c.Assert(lex.Current(), Equals, "jumps")
	lex.FindRune(' ')
	c.Assert(lex.Current(), Equals, "")
	lex.FindRune(' ')
	c.Assert(lex.Current(), Equals, "")
	lex.Find("lazy")
	c.Assert(lex.Current(), Equals, "lazy")

	fmt.Printf("1[%s] '%s' '%d'\n", lex.input[lex.start:lex.pos], lex.input[lex.start:], lex.input[lex.pos])
	fmt.Printf("2[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.input[lex.start], lex.input[lex.pos])
	fmt.Printf("3[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.input[lex.start], lex.input[lex.pos])
	fmt.Printf("4[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.input[lex.start], lex.input[lex.pos])
	fmt.Printf("5[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.input[lex.start], lex.input[lex.pos])
	fmt.Printf("6[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.input[lex.start], lex.input[lex.pos])
	fmt.Printf("7[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.input[lex.start], lex.input[lex.pos])
	fmt.Printf("8[%s] %d %d\n", lex.input[lex.start:lex.pos], lex.start, lex.pos)
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
