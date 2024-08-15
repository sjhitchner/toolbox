package graphviz

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Subgraph struct {
	ID         string
	Label      string
	Attributes map[string]interface{}
	Nodes      []*Node
	Edges      []*Edge
	Subgraphs  []*Subgraph
}

func (t Subgraph) String() string {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (t Subgraph) Dot(indent int, prefix string) string {
	prefix = Prefix(prefix, t.ID)

	var sb strings.Builder
	sb.WriteString("\n" + Indent(indent))
	sb.WriteString(fmt.Sprintf("subgraph cluster_%s {\n", prefix))

	sb.WriteString(Indent(indent + 3))
	sb.WriteString(fmt.Sprintf("label=\"%s\";\n", strings.Title(t.Label)))

	for k, v := range t.Attributes {
		sb.WriteString(Indent(indent + 3))
		sb.WriteString(fmt.Sprintf("%s=%s;\n", k, Quote(v)))
	}

	if len(t.Nodes) > 0 {
		sb.WriteString("\n")
	}
	for _, n := range t.Nodes {
		sb.WriteString(n.Dot(indent+1, prefix) + "\n")
	}

	for _, sg := range t.Subgraphs {
		sb.WriteString(sg.Dot(indent+1, prefix) + "\n")
	}

	if len(t.Edges) > 0 {
		sb.WriteString("\n")
	}
	for _, e := range t.Edges {
		sb.WriteString(e.Dot(indent+1, prefix) + "\n")
	}

	sb.WriteString(Indent(indent))
	sb.WriteString("};")
	return sb.String()
}

func (t *Subgraph) AddNode(node ...*Node) {
	t.Nodes = append(t.Nodes, node...)
}

func (t *Subgraph) AddSubgraph(subsg ...*Subgraph) {
	t.Subgraphs = append(t.Subgraphs, subsg...)
}

func (t *Subgraph) Forward(from, to string) {
	t.Connect(from, to, Forward)
}

func (t *Subgraph) Connect(from, to string, dir Direction) {
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

	if dir == Both {
		edge.Attributes["dir"] = "both"
	}

	t.Edges = append(t.Edges, edge)
}

func (t *Subgraph) findNodeByDotPath(dotPath string) (string, string, error) {
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
