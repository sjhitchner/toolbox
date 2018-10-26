package flag

import (
	"flag"
	. "gopkg.in/check.v1"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	TestingT(t)
}

type LibSuite struct{}

var _ = Suite(&LibSuite{})

func (s *LibSuite) TestPrint(c *C) {
	PrintEnv(os.Stdout)
}

func (s *LibSuite) TestToEnvKey(c *C) {
	flag := "steve-is-awesome"
	env := "STEVE_IS_AWESOME"
	got := toEnvKey(flag)
	c.Assert(got, Equals, env)
}

func (s *LibSuite) TestEnvVariablesString(c *C) {
	flag.Parse()

	if err := os.Setenv("TEST_STRING", "hello"); err != nil {
		c.Fatal(err)
	}
	var str string
	StringVar(&str, "test-string", "world", "")
	c.Assert(str, Equals, "hello")

	StringVar(&str, "test-string-empty", "world", "")
	c.Assert(str, Equals, "world")
}

func (s *LibSuite) TestEnvVariablesFloat64(c *C) {
	flag.Parse()

	if err := os.Setenv("TEST_FLOAT", "123.12"); err != nil {
		c.Fatal(err)
	}
	var f float64
	Float64Var(&f, "test-float", 0, "")
	c.Assert(f, Equals, 123.12)

	if err := os.Setenv("TEST_FLOAT_BAD", "hello"); err != nil {
		c.Fatal(err)
	}
	Float64Var(&f, "test-float-bad", 456, "")
	c.Assert(f, Equals, 456.0)
}

func (s *LibSuite) TestEnvVariablesInt(c *C) {
	flag.Parse()

	if err := os.Setenv("TEST_INT", "123"); err != nil {
		c.Fatal(err)
	}
	var i int
	IntVar(&i, "test-int", 0, "")
	c.Assert(i, Equals, 123)

	if err := os.Setenv("TEST_INT_BAD", "hello"); err != nil {
		c.Fatal(err)
	}
	IntVar(&i, "test-int-bad", 456, "")
	c.Assert(i, Equals, 456)
}

func (s *LibSuite) TestEnvVariablesDuration(c *C) {
	flag.Parse()

	if err := os.Setenv("TEST_DURATION", "123"); err != nil {
		c.Fatal(err)
	}
	var d time.Duration
	DurationVar(&d, "test-duration", 0, "")
	c.Assert(d, Equals, time.Duration(123))

	if err := os.Setenv("TEST_DURATION_BAD", "hello"); err != nil {
		c.Fatal(err)
	}
	DurationVar(&d, "test-duration-bad", 456, "")
	c.Assert(d, Equals, time.Duration(456))
}

func (s *LibSuite) TestEnvVariablesBool(c *C) {
	flag.Parse()

	if err := os.Setenv("TEST_BOOL_TRUE", "true"); err != nil {
		c.Fatal(err)
	}
	var bt bool
	BoolVar(&bt, "test-bool-true", false, "")
	c.Assert(bt, Equals, true)

	if err := os.Setenv("TEST_BOOL_FALSE", "false"); err != nil {
		c.Fatal(err)
	}
	var bf bool
	BoolVar(&bf, "test-bool-false", true, "")
	c.Assert(bf, Equals, false)
}
