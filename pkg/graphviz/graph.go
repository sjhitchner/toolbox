package graphviz

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

type Graph struct {
	ID          string
	ShowLegend  bool
	IsStrict    bool
	IsDigraph   bool
	Label       string
	Attributes  map[string]interface{}
	Global      map[string]map[string]interface{}
	Nodes       []*Node
	Edges       []*Edge
	Subgraphs   []*Subgraph
	ImageMapper ImageMapFn
}

func (t *Graph) AddNode(nodes ...*Node) *Graph {
	t.Nodes = append(t.Nodes, nodes...)
	return t
}

func (t *Graph) AddSubgraph(sgs ...*Subgraph) *Graph {
	t.Subgraphs = append(t.Subgraphs, sgs...)
	return t
}

func (t Graph) String() string {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (t Graph) Dot() string {
	legendCh = make(chan Image, 100)

	var sb strings.Builder
	if t.IsStrict {
		sb.WriteString("strict ")
	}
	if t.IsDigraph {
		sb.WriteString("digraph ")
	} else {
		sb.WriteString("graph ")
	}
	sb.WriteString(fmt.Sprintf("\"%s\" {\n", t.ID))

	// Graph attributes
	for k, v := range t.Attributes {
		sb.WriteString(Indent(2))
		sb.WriteString(fmt.Sprintf("%s=%s;\n", k, Quote(v)))
	}

	// Global attributes
	for k, m := range t.Global {
		sb.WriteString(Indent(2))
		sb.WriteString(k)
		sb.WriteString(" ")
		sb.WriteString(Attributes(m))
		sb.WriteString("\n")
	}

	prefix := ""

	// Nodes
	if len(t.Nodes) > 0 {
		sb.WriteString("\n")
	}
	for _, n := range t.Nodes {
		n.SetImageMapper(t.ImageMapper)
		sb.WriteString(n.Dot(1, prefix) + "\n")
	}

	// Subgraphs
	for _, sg := range t.Subgraphs {
		sg.SetImageMapper(t.ImageMapper)
		sb.WriteString(sg.Dot(1, prefix) + "\n")
	}

	// Edges
	if len(t.Edges) > 0 {
		sb.WriteString("\n")
	}
	for _, e := range t.Edges {
		sb.WriteString(e.Dot(1, prefix) + "\n")
	}

	// Legend
	close(legendCh)
	if t.ShowLegend {
		sb.WriteString("\n  subgraph cluster_legend {\n")
		sb.WriteString("      label=\"Legend\";\n")
		sb.WriteString("      style=solid;\n")
		sb.WriteString("      color=grey;\n")
		//sb.WriteString("      nodesep=0.1;\n")
		//sb.WriteString("      ranksep=0.2;\n")
		for legend := range OrderLegend(legendCh) {
			sb.WriteString(fmt.Sprintf("    Legend%s [label=\"%s\", shape=none, labelloc=b, image=\"%s\"];\n", strcase.ToCamel(legend.Label), legend.Label, legend.Path))
		}
		sb.WriteString("  };\n")
	}

	sb.WriteString("}")
	return sb.String()
}

func (t *Graph) Connect(from, to string, dir ...Direction) *Graph {
	fromNode, tail, err := t.findNodeByDotPath(from)
	if err != nil {
		panic(fmt.Errorf("error finding 'from' node: %v", err))
	}

	toNode, head, err := t.findNodeByDotPath(to)
	if err != nil {
		panic(fmt.Errorf("error finding 'to' node: %v", err))
	}

	edge := &Edge{
		From:       fromNode,
		To:         toNode,
		Tail:       tail,
		Head:       head,
		Attributes: make(map[string]interface{}),
	}

	if len(dir) > 0 {
		switch dir[0] {
		case DirBoth:
			edge.Attributes["dir"] = "both"
		case DirBack:
			edge.Attributes["dir"] = "both"
		}
	}

	t.Edges = append(t.Edges, edge)
	return t
}

func (t *Graph) findNodeByDotPath(dotPath string) (string, string, error) {
	if dotPath == "" {
		return "", "", fmt.Errorf("invalid dot path: %s", dotPath)
	}

	parts := strings.Split(dotPath, ".")
	firstPart := parts[0]
	remainingPath := strings.Join(parts[1:], ".")

	if len(parts) == 1 {
		for _, node := range t.Nodes {
			if node.ID == firstPart {
				return node.ID, "", nil
			}
		}

		for _, sub := range t.Subgraphs {
			if sub.ID == firstPart {
				if len(sub.Nodes) == 0 {
					return "", "", fmt.Errorf("No node in subgraph '%s' to connect", firstPart)
				}

				node := sub.Nodes[0]
				return firstPart + "_" + node.ID, firstPart, nil
			}
		}

		return "", "", fmt.Errorf("subgraph '%s' not found in subgraph '%s'", firstPart, t.ID)
	}

	for _, sub := range t.Subgraphs {
		if sub.ID == firstPart {
			path, sub, err := sub.findNodeByDotPath(remainingPath)
			if sub != "" {
				sub = firstPart + "_" + sub
			}
			return firstPart + "_" + path, sub, err
		}
	}

	return "", "", fmt.Errorf("node '%s' not found in subgraph '%s'", firstPart, t.ID)
}
