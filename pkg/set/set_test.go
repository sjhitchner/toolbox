package set

import (
	"encoding/json"
	"testing"

	. "github.com/sjhitchner/toolbox/pkg/testing"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type SetSuite struct{}

var _ = Suite(&SetSuite{})

func (s *SetSuite) TestAdd(c *C) {
	s1 := New[string]("test1", "test2")
	s1.Add("test3")
	s1.Add("test4", "test5")
	s1.Merge(New[string]("test6", "test7"))
	c.Assert(s1.CheckAndAdd("test7"), Equals, false)
	c.Assert(s1.CheckAndAdd("test8"), Equals, true)

	c.Assert(s1.Cardinality(), Equals, 8)
	c.Assert(s1.Contains("test1"), Equals, true)
	c.Assert(s1.Contains("test2"), Equals, true)
	c.Assert(s1.Contains("test3"), Equals, true)
	c.Assert(s1.Contains("test4"), Equals, true)
	c.Assert(s1.Contains("test5"), Equals, true)
	c.Assert(s1.Contains("test6"), Equals, true)
	c.Assert(s1.Contains("test7"), Equals, true)
	c.Assert(s1.Contains("test8"), Equals, true)
	c.Assert(s1.Contains("test9"), Equals, false)
}

func (s *SetSuite) Test_Cardinality(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")

	c.Assert(s1.Cardinality(), Equals, 2)
}

func (s *SetSuite) Test_Remove(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Remove("test2")

	c.Assert(s1.Contains("test1"), Equals, true)
	c.Assert(s1.Contains("test2"), Equals, false)
	c.Assert(s1.Contains("test3"), Equals, false)
}

func (s *SetSuite) Test_Clear(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")

	c.Assert(s1.Contains("test1"), Equals, true)
	c.Assert(s1.Contains("test2"), Equals, true)

	s1.Clear()

	c.Assert(s1.IsEmpty(), Equals, true)
	c.Assert(s1.Contains("test1"), Equals, false)
	c.Assert(s1.Contains("test2"), Equals, false)
}

func (s *SetSuite) Test_Equal(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")
	s1.Add("test4")
	s1.Add("test5")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")

	s3 := New[string]()
	s3.Add("test3")
	s3.Add("test4")

	c.Assert(s2.Equals(s3), Equals, false)
	c.Assert(s3.Equals(s2), Equals, false)

	c.Assert(s1.Equals(s2), Equals, false)
	c.Assert(s2.Equals(s1), Equals, false)

	s1.Remove("test4")
	s1.Remove("test5")
	s2.Add("test3")

	c.Assert(s1.Equals(s2), Equals, true)
	c.Assert(s2.Equals(s1), Equals, true)

}

func (s *SetSuite) Test_Clone(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")

	s2 := s1.Clone()

	c.Assert(s1.Cardinality(), Equals, s2.Cardinality())
	c.Assert(s2.Equals(s2), Equals, true)
	c.Assert(s2.Equals(s1), Equals, true)
}

func (s *SetSuite) Test_Difference(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")

	sd1 := s1.Difference(s2)
	c.Assert(sd1.Cardinality(), Equals, 1)
	c.Assert(sd1.Contains("test1"), Equals, false)
	c.Assert(sd1.Contains("test2"), Equals, false)
	c.Assert(sd1.Contains("test3"), Equals, true)

	sd2 := s2.Difference(s1)
	c.Assert(sd2.Cardinality(), Equals, 0)
	c.Assert(sd2.IsEmpty(), Equals, true)
	c.Assert(sd2.Contains("test1"), Equals, false)
	c.Assert(sd2.Contains("test2"), Equals, false)
	c.Assert(sd2.Contains("test3"), Equals, false)
}

func (s *SetSuite) Test_Intersect(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")
	s1.Add("test5")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")
	s2.Add("test4")

	si1 := s1.Intersect(s2)
	c.Assert(si1.Cardinality(), Equals, 2)
	c.Assert(si1.Contains("test1"), Equals, true)
	c.Assert(si1.Contains("test2"), Equals, true)
	c.Assert(si1.Contains("test3"), Equals, false)
	c.Assert(si1.Contains("test4"), Equals, false)
	c.Assert(si1.Contains("test5"), Equals, false)

	si2 := s2.Intersect(s1)
	c.Assert(si2.Cardinality(), Equals, 2)
	c.Assert(si2.Contains("test1"), Equals, true)
	c.Assert(si2.Contains("test2"), Equals, true)
	c.Assert(si2.Contains("test3"), Equals, false)
	c.Assert(si2.Contains("test4"), Equals, false)
	c.Assert(si1.Contains("test5"), Equals, false)
}

func (s *SetSuite) Test_Union(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")
	s2.Add("test4")

	su1 := s1.Union(s2)
	c.Assert(su1.Cardinality(), Equals, 4)
	c.Assert(su1.Contains("test1"), Equals, true)
	c.Assert(su1.Contains("test2"), Equals, true)
	c.Assert(su1.Contains("test3"), Equals, true)
	c.Assert(su1.Contains("test4"), Equals, true)

	su2 := s2.Union(s1)
	c.Assert(su2.Cardinality(), Equals, 4)
	c.Assert(su2.Contains("test1"), Equals, true)
	c.Assert(su2.Contains("test2"), Equals, true)
	c.Assert(su2.Contains("test3"), Equals, true)
	c.Assert(su2.Contains("test4"), Equals, true)
}

func (s *SetSuite) Test_IsSubset(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")

	c.Assert(s1.IsSubset(s2), Equals, false)
	c.Assert(s2.IsSubset(s1), Equals, true)
}

func (s *SetSuite) Test_IsSuperset(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")

	c.Assert(s1.IsSuperset(s2), Equals, true)
	c.Assert(s2.IsSuperset(s1), Equals, false)
}

func (s *SetSuite) Test_Iter(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	s2 := New[string]()
	for s := range s1.Iterator().C {
		s2.Add(s)
	}
	c.Assert(s1.Equals(s2), Equals, true)

	s3 := New[string]()
	for s := range s1.Iter() {
		s3.Add(s)
	}
	c.Assert(s1.Equals(s3), Equals, true)
}

func (s *SetSuite) Test_String(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	str := "Set{test1, test2, test3}"
	c.Assert(len(s1.String()), Equals, len(str))
}

func (s *SetSuite) Test_SymmetricDifference(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test2")
	s2.Add("test4")

	sd1 := s1.SymmetricDifference(s2)
	c.Assert(sd1.Cardinality(), Equals, 2)
	c.Assert(sd1.Contains("test1"), Equals, false)
	c.Assert(sd1.Contains("test2"), Equals, false)
	c.Assert(sd1.Contains("test3"), Equals, true)
	c.Assert(sd1.Contains("test4"), Equals, true)

	sd2 := s2.SymmetricDifference(s1)
	c.Assert(sd2.Cardinality(), Equals, 2)
	c.Assert(sd2.Contains("test1"), Equals, false)
	c.Assert(sd2.Contains("test2"), Equals, false)
	c.Assert(sd2.Contains("test3"), Equals, true)
	c.Assert(sd2.Contains("test4"), Equals, true)
}

func (s *SetSuite) Test_HasOverlap(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")
	s1.Add("test4")

	s2 := New[string]()
	s2.Add("test1")
	s2.Add("test4")
	s2.Add("test5")

	c.Assert(s1.HasOverlap(s2), Equals, true)
	c.Assert(s2.HasOverlap(s1), Equals, true)

	s1.Remove("test1")
	s2.Remove("test1")

	c.Assert(s1.HasOverlap(s2), Equals, true)
	c.Assert(s2.HasOverlap(s1), Equals, true)

	s2.Add("test6")
	s2.Add("test7")

	c.Assert(s1.HasOverlap(s2), Equals, true)
	c.Assert(s2.HasOverlap(s1), Equals, true)

	s1.Remove("test4")
	s2.Remove("test4")

	c.Assert(s1.HasOverlap(s2), Equals, false)
	c.Assert(s2.HasOverlap(s1), Equals, false)

	s2.Add("test7")
	s2.Add("test8")
	s2.Add("test9")

	c.Assert(s1.HasOverlap(s2), Equals, false)
	c.Assert(s2.HasOverlap(s1), Equals, false)
}

func (s *SetSuite) Test_ToSlice(c *C) {
	s1 := New[string]()
	s1.Add("test1")
	s1.Add("test2")
	s1.Add("test3")

	slice := s1.ToSlice()

	c.Assert(slice, HasLen, 3)
	c.Assert(slice, Contains, "test1")
	c.Assert(slice, Contains, "test2")
	c.Assert(slice, Contains, "test3")
}

func (s *SetSuite) Test_JSONMarshaling(c *C) {
	data := []string{
		"test1",
		"test2",
		"test3",
		"test4",
	}

	expected := New[string](data...)

	serialized, err := json.Marshal(expected)
	c.Assert(err, IsNil, Commentf("Marshalled (%s)", serialized))

	actual := New[string]()
	err = json.Unmarshal(serialized, &actual)
	c.Assert(err, IsNil)

	c.Assert(expected.Equals(actual), IsTrue)
}
