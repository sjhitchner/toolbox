package streaming

import (
	"sync"
	"testing"
	"time"

	// . "github.com/sjhitchner/toolbox/pkg/testing"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type StreamingSuite struct{}

var _ = Suite(&StreamingSuite{})

func (s *StreamingSuite) Test_Generate_Nil(c *C) {
	in := Generate(nil, 1, 2, 3, 4, 5)

	var index int
	for _ = range in {
		index++
	}
	c.Assert(index, Equals, 5)
}

func (s *StreamingSuite) Test_Generate_Done(c *C) {
	done := make(chan struct{})

	in := Generate(done, 1, 2, 3, 4, 5)

	<-in
	close(done)

	for _ = range in {
		c.Fail()
	}
}

func (s *StreamingSuite) Test_Apply(c *C) {

	in := Generate(nil, 1, 2, 3, 4, 5)

	s2 := func(i int) int {
		return i * i
	}

	out := Apply[int](nil, in, s2)

	for i := 1; i < 6; i++ {
		expected := i * i
		obtained := <-out

		c.Assert(obtained, Equals, expected)
	}
}

func (s *StreamingSuite) Test_Apply_Done(c *C) {
	done := make(chan struct{})

	in := Generate(nil, 1, 2, 3, 4, 5)

	s2 := func(i int) int {
		return i * i
	}

	out := Apply[int](done, in, s2)

	close(done)
	<-out
	for _ = range out {
		c.Fail()
	}
}

func (s *StreamingSuite) Test_Merge(c *C) {
	done := make(chan struct{})

	in1 := Generate(done, 1, 2, 3)
	in2 := Generate(done, 4, 5, 6)

	out := Merge[int](done, in1, in2)

	var index int
	for _ = range out {
		index++
	}

	c.Assert(index, Equals, 6)
}

func (s *StreamingSuite) Test_Merge_Done(c *C) {
	done := make(chan struct{})

	in1 := Generate(nil, 1, 2, 3)
	in2 := Generate(nil, 4, 5, 6)

	out := Merge[int](done, in1, in2)

	close(done)

	<-out
	for _ = range out {
		c.Fail()
	}
}

func (s *StreamingSuite) Test_Multiplex(c *C) {
	done := make(chan struct{})

	source := Generate(nil, 1, 2, 3)

	out1 := make(chan int)
	out2 := make(chan int)

	var wg sync.WaitGroup

	Multiplex[int](done, source, out1, out2)

	read := func(ch <-chan int) {
		defer wg.Done()

		index := 1
		for v := range ch {
			c.Assert(v, Equals, index)
			index++
		}
	}

	wg.Add(2)
	go read(out1)
	go read(out2)

	wg.Wait()
}

func (s *StreamingSuite) Test_Multiplex_Done(c *C) {
	done := make(chan struct{})

	source := Generate(done, 1, 2, 3)

	out1 := make(chan int)
	out2 := make(chan int)

	var wg sync.WaitGroup

	Multiplex[int](done, source, out1, out2)

	read := func(ch <-chan int) {
		defer wg.Done()

		for _ = range ch {
			c.Fail()
		}
	}

	wg.Add(2)

	close(done)

	go read(out1)
	go read(out2)

	wg.Wait()
}

func (s *StreamingSuite) Test_Or(c *C) {

	sig := func(after time.Duration) <-chan struct{} {
		c := make(chan struct{})
		go func() {
			defer close(c)
			<-time.After(after)
		}()
		return c
	}

	<-Done(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(100*time.Millisecond),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	<-Done(
		sig(2*time.Hour),
		sig(100*time.Millisecond),
		sig(5*time.Minute),
		sig(1*time.Minute),
	)

	<-Done(
		sig(100*time.Millisecond),
		sig(5*time.Minute),
		sig(1*time.Minute),
	)

	<-Done(
		sig(2*time.Hour),
		sig(100*time.Millisecond),
	)

	<-Done(
		sig(100 * time.Millisecond),
	)

	<-Done(
		nil,
		sig(100*time.Millisecond),
	)
}
