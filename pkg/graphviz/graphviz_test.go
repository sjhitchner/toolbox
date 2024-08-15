package graphviz

import (
	"fmt"
	"testing"

	"gopkg.in/check.v1"
	// Adjust the import path as needed
)

// ... (Your existing code, including the Graph, Node, Edge, and Subgraph structs)

// Suite struct for gocheck
type DotSuite struct {
	subgraph *Subgraph
}

var _ = check.Suite(&DotSuite{})

func Test(t *testing.T) {
	check.TestingT(t)
}

func (s *DotSuite) SetUpTest(c *check.C) {
	s.subgraph = &Subgraph{
		ID: "root",
		Nodes: []*Node{
			{ID: "node1"},
			{ID: "node2"},
		},
		Subgraphs: []*Subgraph{
			{
				ID: "subgraph1",
				Nodes: []*Node{
					{ID: "node3"},
				},
				Subgraphs: []*Subgraph{
					{
						ID: "subgraph2",
						Nodes: []*Node{
							{ID: "node4"},
						},
					},
				},
			},
		},
	}
	// fmt.Println(s.subgraph)
	// fmt.Println(s.subgraph.Dot(0, ""))

}

func (s *DotSuite) TearDownTest(c *check.C) {
	//fmt.Println(s.subgraph)
	fmt.Println(s.subgraph.Dot(0, ""))
}

// Test case for successful connection
func (s *DotSuite) TestConnectSuccess(c *check.C) {
	s.subgraph.Forward("node1", "node2")
	fmt.Println(s.subgraph)
	c.Assert(len(s.subgraph.Edges), check.Equals, 1)
	c.Assert(s.subgraph.Edges[0].From, check.Equals, "node1")
	c.Assert(s.subgraph.Edges[0].To, check.Equals, "node2")

	s.subgraph.Forward("node1", "subgraph1.node3")
	c.Assert(len(s.subgraph.Edges), check.Equals, 2)
	c.Assert(s.subgraph.Edges[1].From, check.Equals, "node1")
	c.Assert(s.subgraph.Edges[1].To, check.Equals, "subgraph1_node3")

	s.subgraph.Forward("node2", "subgraph1.subgraph2.node4")
	c.Assert(len(s.subgraph.Edges), check.Equals, 3)
	c.Assert(s.subgraph.Edges[2].From, check.Equals, "node2")
	c.Assert(s.subgraph.Edges[2].To, check.Equals, "subgraph1_subgraph2_node4")

	s.subgraph.Forward("subgraph1.subgraph2.node4", "node2")
	c.Assert(len(s.subgraph.Edges), check.Equals, 4)
	c.Assert(s.subgraph.Edges[3].From, check.Equals, "subgraph1_subgraph2_node4")
	c.Assert(s.subgraph.Edges[3].To, check.Equals, "node2")
}

func (s *DotSuite) TestConnectSuccessHead(c *check.C) {
	s.subgraph.Forward("node1", "subgraph1")
	c.Assert(len(s.subgraph.Edges), check.Equals, 1)
	c.Assert(s.subgraph.Edges[0].From, check.Equals, "node1")
	c.Assert(s.subgraph.Edges[0].To, check.Equals, "subgraph1_node3")

	s.subgraph.Forward("node1", "subgraph1.subgraph2")
	c.Assert(len(s.subgraph.Edges), check.Equals, 2)
	c.Assert(s.subgraph.Edges[1].From, check.Equals, "node1")
	c.Assert(s.subgraph.Edges[1].To, check.Equals, "subgraph1_subgraph2_node4")
}

/*
// Test case for node not found error
func (s *DotSuite) TestConnectNodeNotFound(c *check.C) {
	// ... (Use the same s.subgraph or create a new one)

	// Try to connect to a non-existent node
	s.subgraph.Connect("node1", "non_existent_node")
	//c.Assert(err.Error(), check.Equals, "error finding 'to' node: node 'non_existent_node' not found in subgraph 'test_subgraph'")
}

// Test case for subgraph not found error
func (s *DotSuite) TestConnectSubgraphNotFound(c *check.C) {
	// ... (Use the same s.subgraph or create a new one)

	// Try to connect to a node within a non-existent subgraph
	s.subgraph.Connect("node1", "non_existent_subgraph.node3")
	//c.Assert(err.Error(), check.Equals, "error finding 'to' node: subgraph 'non_existent_subgraph' not found in subgraph 'test_subgraph'")
}

// Test case for invalid dot path
func (s *DotSuite) TestConnectInvalidDotPath(c *check.C) {
	// ... (Use the same s.subgraph or create a new one)

	// Provide an invalid dot path
	s.subgraph.Connect("node1", "invalid..path")
	//c.Assert(err, check.NotNil)
	//c.Assert(err.Error(), check.Equals, "error finding 'to' node: invalid dot path: invalid..path")
}
*/
