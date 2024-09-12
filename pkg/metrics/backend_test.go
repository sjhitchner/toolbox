package metrics

import (
	"math/rand"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MetricSuite struct {
}

var _ = Suite(&MetricSuite{})

func (s *MetricSuite) SetUpTest(c *C) {
}

func (s *MetricSuite) Benchmark_Counter(c *C) {
	for i := 0; i < c.N; i++ {
		CounterAt("key", rand.Intn(1000)).Emit()
	}
}

func (s *MetricSuite) Benchmark_Timer(c *C) {
	for i := 0; i < c.N; i++ {
		Timer("key").Emit()
	}
}

/*
	pool := &sync.Pool{
		New: func() any { return make([]byte, 64) },
	}

	for i := 0; i < b.N; i++ {
		pool.Put(make([]byte, 64))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Get()
	}
}

func BenchmarkSyncPoolFill(b *testing.B) {
	pool := &sync.Pool{
		New: func() any { return make([]byte, 64) },
	}

	for i := 0; i < b.N; i++ {
		pool.Put(make([]byte, 64))
	}
}

/*
func (s *MetricSuite) TestNewClient(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)
	c.Assert(m.(*client).Conn, NotNil)
}

func (s *MetricSuite) TestConnectInvalidAddress(c *C) {
	m := NewNetworkClient("invalid", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, ErrorMatches, "Error resolving statsite:.*")
}

func (s *MetricSuite) TestNewClientConnect(c *C) {
	m := NewNetworkClient("badconnection", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, ErrorMatches, "Error connecting to statsite:.*")
}

func (s *MetricSuite) TestEmit(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)

	msg := NewKeyValue("key", "value")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	err = m.Emit(msg)
	c.Assert(err, IsNil)
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 1)
	c.Assert(s.mockStatsite.Last(), Equals, msg.String())
}

func (s *MetricSuite) TestEmitMultiple(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)

	msg := NewKeyValue("key", "value")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	for i := 0; i < 10; i++ {
		err := m.Connect()
		c.Assert(err, IsNil)
		err = m.Emit(msg)
		c.Assert(err, IsNil)
	}
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 10)
	c.Assert(s.mockStatsite.Last(), Equals, msg.String())
}

func (s *MetricSuite) TestEmitDouble(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)

	msg := NewKeyValue("key", "value")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	err = m.Emit(msg)
	c.Assert(err, IsNil)
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 1)
	m.Emit(msg)
	c.Assert(s.mockStatsite.Count(), Equals, 2)
}

func (s *MetricSuite) TestEmitNotConnected(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	conn := m.(*client).Conn
	c.Assert(conn, IsNil)
	msg := NewKeyValue("key", "value")
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	err := m.Emit(msg)
	c.Assert(err, IsNil)
	conn = m.(*client).Conn
	c.Assert(conn, NotNil)
	// Expect 1 stat added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 1)
}

func (s *MetricSuite) TestEmitFailure(c *C) {
	m := NewNetworkClient("statsite", s.mockNetwork)
	err := m.Connect()
	c.Assert(err, IsNil)
	msg := NewKeyValue("bad", "key")
	// test 0 sent
	c.Assert(s.mockStatsite.Count(), Equals, 0)
	// Expect error
	err = m.Emit(msg)
	c.Assert(err, NotNil)
	// Expect no stats added to statsite
	c.Assert(s.mockStatsite.Count(), Equals, 0)
}
*/
