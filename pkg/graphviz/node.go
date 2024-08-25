package graphviz

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

type Node struct {
	ID         string
	Label      string
	Image      ImageType
	Attributes map[string]interface{}
}

func (t Node) Dot(indent int, prefix string) string {
	prefix = Prefix(prefix, t.ID)

	attrs := []string{
		fmt.Sprintf("label=%s", Quote(t.Label)),
	}
	if t.Image > 0 {
		attrs = append(attrs,
			"shape=none",
			"labelloc=m",
			fmt.Sprintf("image=\"%s\"", ImageMapper(prefix, t.Image)),
		)
	}

	var sb strings.Builder
	sb.WriteString(Indent(indent))
	sb.WriteString(strcase.ToCamel(prefix))
	sb.WriteString(" ")
	sb.WriteString(Attributes(t.Attributes, attrs...))
	return sb.String()
}

func (t Node) String() string {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
