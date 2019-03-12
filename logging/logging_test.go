package logging

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type LoggingSuite struct {
	ts *httptest.Server
}

var _ = Suite(&LoggingSuite{})

func (s *LoggingSuite) SetUpSuite(c *C) {
	s.ts = httptest.NewServer(http.HandlerFunc(LoggingHandler))
	log.SetLevel(log.InfoLevel)
}

func (s *LoggingSuite) TearDownSuite(c *C) {
	defer s.ts.Close()
}

func (s *LoggingSuite) Test_GET(c *C) {
	url := fmt.Sprintf("%s/logging", s.ts.URL)

	resp, err := http.Get(url)
	c.Assert(err, IsNil)

	var lr LoggingResponse
	err = json.NewDecoder(resp.Body).Decode(&lr)
	c.Assert(err, IsNil)

	c.Assert(lr.Status, Equals, http.StatusOK)
	c.Assert(lr.Level, Equals, log.InfoLevel.String())
	c.Assert(lr.Error, Equals, "")
}

func (s *LoggingSuite) Test_POST(c *C) {
	url := fmt.Sprintf("%s/logging", s.ts.URL)

	{
		var lr LoggingResponse
		resp, err := http.Post(url+"?level=invalid", "", nil)
		c.Assert(err, IsNil)
		err = json.NewDecoder(resp.Body).Decode(&lr)
		c.Assert(err, IsNil)
		c.Assert(lr.Status, Equals, http.StatusBadRequest)
		c.Assert(lr.Error, Not(Equals), "")
	}
	{
		var lr LoggingResponse
		resp, err := http.Get(url)
		c.Assert(err, IsNil)
		err = json.NewDecoder(resp.Body).Decode(&lr)
		c.Assert(err, IsNil)
		c.Assert(lr.Status, Equals, http.StatusOK)
		c.Assert(lr.Level, Equals, log.InfoLevel.String())
		c.Assert(lr.Error, Equals, "")
	}
	{
		var lr LoggingResponse
		resp, err := http.Post(url+"?level=error", "", nil)
		c.Assert(err, IsNil)
		err = json.NewDecoder(resp.Body).Decode(&lr)
		c.Assert(err, IsNil)
		c.Assert(lr.Status, Equals, http.StatusOK)
		c.Assert(lr.Level, Equals, log.ErrorLevel.String())
		c.Assert(lr.Error, Equals, "")
	}
	{
		var lr LoggingResponse
		resp, err := http.Get(url)
		c.Assert(err, IsNil)
		err = json.NewDecoder(resp.Body).Decode(&lr)
		c.Assert(err, IsNil)
		c.Assert(lr.Status, Equals, http.StatusOK)
		c.Assert(lr.Level, Equals, log.ErrorLevel.String())
		c.Assert(lr.Error, Equals, "")
	}
	{
		var lr LoggingResponse
		resp, err := http.Post(url+"?level=info", "", nil)
		c.Assert(err, IsNil)
		err = json.NewDecoder(resp.Body).Decode(&lr)
		c.Assert(err, IsNil)
		c.Assert(lr.Status, Equals, http.StatusOK)
		c.Assert(lr.Level, Equals, log.InfoLevel.String())
		c.Assert(lr.Error, Equals, "")
	}
	{
		var lr LoggingResponse
		resp, err := http.Get(url)
		c.Assert(err, IsNil)
		err = json.NewDecoder(resp.Body).Decode(&lr)
		c.Assert(err, IsNil)
		c.Assert(lr.Status, Equals, http.StatusOK)
		c.Assert(lr.Level, Equals, log.InfoLevel.String())
		c.Assert(lr.Error, Equals, "")
	}
}
