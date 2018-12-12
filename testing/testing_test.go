package testing

import (
	gocheck "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) {
	gocheck.TestingT(t)
}

type TestingSuite struct{}

var _ = gocheck.Suite(&TestingSuite{})

func (s *TestingSuite) Test_IsTrue(c *gocheck.C) {
	c.Assert(true, IsTrue)
}

func (s *TestingSuite) Test_IsFalse(c *gocheck.C) {
	c.Assert(false, IsFalse)
}

func (s *TestingSuite) Test_EqualsWithin(c *gocheck.C) {
	c.Assert(10.0, EqualsWithin, 10.1, 0.1)
}

func (s *TestingSuite) Test_Between(c *gocheck.C) {
	c.Assert(10, Between, 9, 11)
	c.Assert(0.5, Between, 0.4, 0.6)
	c.Assert(int64(10), Between, int64(9), int64(11))
	c.Assert(float32(10), Between, float32(9), float32(11))
	c.Assert(float64(10), Between, float64(9), float64(11))
}

func (s *TestingSuite) Test_Contains(c *gocheck.C) {
	c.Assert("hello, world!", Contains, "hello")
	c.Assert([]int{1, 2, 3, 4}, Contains, 1)
	c.Assert([]int64{1, 2, 3, 4}, Contains, int64(1))
	c.Assert([]float32{1, 2, 3, 4}, Contains, float32(1))
	c.Assert([]float64{1, 2, 3, 4}, Contains, float64(1))
	c.Assert([]string{"1", "2", "3", "4"}, Contains, "1")
}
