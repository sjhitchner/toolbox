package csv

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const ()

func Test(t *testing.T) {
	TestingT(t)
}

type CSVSuite struct {
}

var _ = Suite(&CSVSuite{})

func (s *CSVSuite) SetUpSuite(c *C) {
}

func (s *CSVSuite) Test_Decode(c *C) {
	reader := csv.NewReader(strings.NewReader(CSV1))

	var rows []Row1
	c.Assert(NewDecoder(reader).Decode(&rows), IsNil)

	c.Assert(rows[0].Foo, Equals, "foo1")
	c.Assert(rows[0].Qwerty, Equals, "qwerty1")
	c.Assert(rows[0].Date.Format(defaultDateFormat), Equals, "2021-01-01T00:00:00Z")

	c.Assert(rows[1].Foo, Equals, "foo2")
	c.Assert(rows[1].Qwerty, Equals, "qwerty2")
	c.Assert(rows[1].Date.Format(defaultDateFormat), Equals, "2021-02-02T00:00:00Z")

	c.Assert(rows[2].Foo, Equals, "foo3")
	c.Assert(rows[2].Qwerty, Equals, "qwerty3")
	c.Assert(rows[2].Date.Format(defaultDateFormat), Equals, "2021-03-03T00:00:00Z")
}

const CSV1 = `foo1,qwerty1,2021-01-01T00:00:00Z
foo2,qwerty2,2021-02-02T00:00:00Z
foo3,qwerty3,2021-03-03T00:00:00Z`

type Row1 struct {
	Foo    string
	Qwerty string
	Date   time.Time
}
