package csv

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

const (
	CSVNoHeader = `foo1,qwerty1,2021-01-01T00:00:00Z
foo2,qwerty2,2021-02-02T00:00:00Z
foo3,qwerty3,2021-03-03T00:00:00Z`

	CSVHeader = "foo,qwerty,date\n" + CSVNoHeader

	TSVNoHeader = `foo1	qwerty1	2021-01-01T00:00:00Z
foo2	qwerty2	2021-02-02T00:00:00Z
foo3	qwerty3	2021-03-03T00:00:00Z`

	TSVHeader = "foo	query	date\n" + TSVNoHeader
)

type Row struct {
	Foo    string
	Qwerty string
	Date   time.Time
}

type ColNumTag struct {
	Foo    string    `csv:"0"`
	Qwerty string    `csv:"1"`
	Date   time.Time `csv:"2"`
}

type ColNameTag struct {
	Foo    string    `csv:"foo"`
	Qwerty string    `csv:"qwerty"`
	Date   time.Time `csv:"date"`
}

type NoTag struct {
	Foo    string
	Qwerty string
	Date   time.Time
}

func Test(t *testing.T) {
	TestingT(t)
}

type CSVSuite struct {
}

var _ = Suite(&CSVSuite{})

func (s *CSVSuite) SetUpSuite(c *C) {
}

func (s *CSVSuite) Test_Decode(c *C) {
	reader := csv.NewReader(strings.NewReader(CSVNoHeader))

	var rows []Row
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

func (s *CSVSuite) Test_Stream_NumTag(c *C) {
	reader := csv.NewReader(strings.NewReader(CSVNoHeader))
	outCh, errCh, err := Stream[ColNumTag](reader, false)
	c.Assert(err, IsNil)

	assertNoError(c, errCh)

	row0 := <-outCh
	c.Assert(row0.Foo, Equals, "foo1")
	c.Assert(row0.Qwerty, Equals, "qwerty1")
	c.Assert(row0.Date.Format(defaultDateFormat), Equals, "2021-01-01T00:00:00Z")

	row1 := <-outCh
	c.Assert(row1.Foo, Equals, "foo2")
	c.Assert(row1.Qwerty, Equals, "qwerty2")
	c.Assert(row1.Date.Format(defaultDateFormat), Equals, "2021-02-02T00:00:00Z")

	row2 := <-outCh
	c.Assert(row2.Foo, Equals, "foo3")
	c.Assert(row2.Qwerty, Equals, "qwerty3")
	c.Assert(row2.Date.Format(defaultDateFormat), Equals, "2021-03-03T00:00:00Z")
}

func (s *CSVSuite) Test_Stream_NameTag(c *C) {
	reader := csv.NewReader(strings.NewReader(CSVHeader))
	outCh, errCh, err := Stream[ColNameTag](reader, true)
	c.Assert(err, IsNil)

	assertNoError(c, errCh)

	row0 := <-outCh
	c.Assert(row0.Foo, Equals, "foo1")
	c.Assert(row0.Qwerty, Equals, "qwerty1")
	c.Assert(row0.Date.Format(defaultDateFormat), Equals, "2021-01-01T00:00:00Z")

	row1 := <-outCh
	c.Assert(row1.Foo, Equals, "foo2")
	c.Assert(row1.Qwerty, Equals, "qwerty2")
	c.Assert(row1.Date.Format(defaultDateFormat), Equals, "2021-02-02T00:00:00Z")

	row2 := <-outCh
	c.Assert(row2.Foo, Equals, "foo3")
	c.Assert(row2.Qwerty, Equals, "qwerty3")
	c.Assert(row2.Date.Format(defaultDateFormat), Equals, "2021-03-03T00:00:00Z")
}

func assertNoError(c *C, errCh <-chan error) {
	go func() {
		for err := range errCh {
			c.Assert(err, IsNil)
		}
	}()
}
