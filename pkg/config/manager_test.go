package config

import (
	"fmt"
	"testing"
	//	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&ConfigSuite{})

// ConfigSuite is the test suite struct for configuration tests.
type ConfigSuite struct{}

// SetUpSuite is called before any test or benchmark runs.
func (s *ConfigSuite) SetUpSuite(c *C) {
	fmt.Println("SetUpSuite: called before all tests")
}

// SetUpTest is called before every test or benchmark runs.
func (s *ConfigSuite) SetUpTest(c *C) {
	fmt.Println("SetUpTest: called before each test")
}

// TearDownSuite is called after all tests or benchmarks have run.
func (s *ConfigSuite) TearDownSuite(c *C) {
	fmt.Println("TearDownSuite: called after all tests")
}

// TearDownTest is called after each test or benchmark runs.
func (s *ConfigSuite) TearDownTest(c *C) {
	fmt.Println("TearDownTest: called after each test")
}

func (s *ConfigSuite) Test_Provider_JSON(c *C) {
	pv, err := NewJSONProviderFromString(
		`{"s": "hello", "i": 1, "f": 1, "b": true}`)
	c.Assert(err, IsNil)

	i, err := pv.GetInt("i")
	c.Assert(err, IsNil)

	i64, err := pv.GetInt64("i")
	c.Assert(err, IsNil)

	f64, err := pv.GetFloat64("f")
	c.Assert(err, IsNil)

	str, err := pv.GetString("s")
	c.Assert(err, IsNil)

	b, err := pv.GetBool("b")
	c.Assert(err, IsNil)

	c.Assert(i, Equals, 1)
	c.Assert(i64, Equals, int64(1))
	c.Assert(f64, Equals, 1.0)
	c.Assert(str, Equals, "hello")
	c.Assert(b, Equals, true)
}

func (s *ConfigSuite) Test_Provider_YAML(c *C) {
	pv, err := NewYAMLProviderFromString(`
s: hello
i: 2
f: 2
b: true`)
	c.Assert(err, IsNil)

	i, err := pv.GetInt("i")
	c.Assert(err, IsNil)

	i64, err := pv.GetInt64("i")
	c.Assert(err, IsNil)

	f64, err := pv.GetFloat64("f")
	c.Assert(err, IsNil)

	str, err := pv.GetString("s")
	c.Assert(err, IsNil)

	b, err := pv.GetBool("b")
	c.Assert(err, IsNil)

	c.Assert(i, Equals, 2)
	c.Assert(i64, Equals, int64(2))
	c.Assert(f64, Equals, 2.0)
	c.Assert(str, Equals, "hello")
	c.Assert(b, Equals, true)
}

/*
func (s *ConfigSuite) Test_Priority1(c *C) {

	mgr := New()

	pv1, err := NewJSONProviderFromString(
		`{"s": "1", "i": 1, "f": 1, "b": true}`)
	c.Assert(err, IsNil)
	pv2, err := NewYAMLProviderFromString(`
	s: 2
	i: 2
	f: 2
	b: true
	`)
	c.Assert(err, IsNil)
	pv3, err := NewJSONProviderFromString(
		`{"s": "3", "i": 3, "f": 3, "b": true}`)
	c.Assert(err, IsNil)

	mgr.AddProvider(pv1, 1, time.Second)
	mgr.AddProvider(pv2, 2, time.Second)
	mgr.AddProvider(pv3, 3, time.Second)

	c.Assert(mgr.GetInt("i", 10), Equals, 1)
	c.Assert(mgr.GetInt64("i", 10), Equals, int64(1))
	c.Assert(mgr.GetFloat64("f", 10), Equals, 1.0)
	c.Assert(mgr.GetString("s", "10"), Equals, "1")
	c.Assert(mgr.GetBool("b", false), Equals, true)
}
*/
