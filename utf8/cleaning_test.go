package utf8

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&Utf8Suite{})

type Utf8Suite struct {
}

const (
	BadText = "This policy explains what information we collect when you use Medium\\u2019s sites, services, mobile applications, products, and content (\\u201cServices\\u201d). It also has information about how we store, use, transfer, and delete that information. Our aim is not just to comply with privacy law. It\\u2019s to earn your trust."

	GoodText = "This policy explains what information we collect when you use Medium’s sites, services, mobile applications, products, and content (‒Services–). It also has information about how we store, use, transfer, and delete that information. Our aim is not just to comply with privacy law. It’s to earn your trust."
)

func (s *Utf8Suite) Test_CleanSource(c *C) {
	cleaned, err := CleanDoubleEscapedUtf8(BadText)
	c.Assert(err, IsNil)
	c.Assert(GoodText, Equals, cleaned)
}
