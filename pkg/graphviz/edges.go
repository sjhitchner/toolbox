package graphviz

import (
	"strings"

	"github.com/iancoleman/strcase"
)

type Edge struct {
	From       string
	To         string
	Head       string
	Tail       string
	Attributes map[string]interface{}
}

func (t Edge) Dot(indent int, prefix string) string {
	from := t.From
	if !strings.HasPrefix(t.From, prefix) {
		from = Prefix(prefix, from)
	}

	to := t.To
	if !strings.HasPrefix(t.To, prefix) {
		to = Prefix(prefix, to)
	}

	var sb strings.Builder
	sb.WriteString(Indent(indent))
	sb.WriteString(strcase.ToCamel(from))
	sb.WriteString(" -> ")
	sb.WriteString(strcase.ToCamel(to))

	attrs := make([]string, 0, 2)
	if t.Tail != "" {
		attrs = append(attrs, "ltail=cluster_"+Prefix(prefix, t.Tail))
	}
	if t.Head != "" {
		attrs = append(attrs, "lhead=cluster_"+Prefix(prefix, t.Head))
	}
	sb.WriteString(Attributes(t.Attributes, attrs...))
	sb.WriteString(";")
	return sb.String()
}
