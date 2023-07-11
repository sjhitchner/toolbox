package future

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type FutureSuite struct{}

var _ = Suite(&FutureSuite{})

func (s *FutureSuite) Test_Future_OK(c *C) {

	ctx := context.Background()

	fut := New[int](ctx, func() (int, error) {
		return 5, nil
	})

	result, err := fut.Wait()
	c.Assert(*result, Equals, 5)
	c.Assert(err, IsNil)
}

func (s *FutureSuite) Test_Future_Error(c *C) {

	ctx := context.Background()

	fut := New[int](ctx, func() (int, error) {
		return 0, fmt.Errorf("error")
	})

	result, err := fut.Wait()
	c.Assert(result, IsNil)
	c.Assert(err, NotNil)
}

func (s *FutureSuite) Test_Future_Cancel(c *C) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		1*time.Second)
	defer cancel()

	fut := New[int](ctx, func() (int, error) {
		<-time.After(2 * time.Second)
		return 5, nil
	})

	result, err := fut.Wait()
	c.Assert(result, IsNil)
	c.Assert(err, NotNil)
}
