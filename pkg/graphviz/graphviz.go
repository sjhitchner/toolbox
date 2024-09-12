/*
Primatives to help easy generate consistent GraphViz diagrams

# Library is not thread-safe

For images you need to define a set of ImageTypes using the ImageType type and a custom ImageMapper function that maps the ImageType to file path.  If the Graph has the ShowLegend set a legend will automatically be added with all the images used by that graph.

	const (
		None gv.ImageType = iota
		RDS
		S3
		SNS
		SQS
	)

	func CustomImageMapper(id string, imageType ImageType) Image {
		switch imageType {
			case RDS:
				return Image{Label: "RDS", Path: "images/rds.png"}
			case S3:
				return Image{Label: "S3", Path: "images/s3.png"}
			case SNS:
				return Image{Label: "SNS", Path: "images/sns.png"}
			case SQS:
				return Image{Label: "SQS", Path: "images/sqs.png"}
			default:
				panic("Invalid ImageType")
		}
	}

	SetImageMapper(CustomImageMapper)
*/
package graphviz

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const Tab = "  "

type Direction string

const (
	DirForward Direction = "forward"
	DirBoth              = "both"
	DirBack              = "back"
)

type ImageType int

var legendCh chan Image

type Image struct {
	Label string
	Path  string
}

func Prefix(prefix, id string) string {
	// id = id + "_"
	if prefix == "" {
		return id
	}
	return prefix + "_" + id
}

func Indent(indent int) string {
	return strings.Repeat(Tab, indent)
}

func Quote(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.3f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return fmt.Sprintf("\"%s\"", v)
	}
	return ""
}

func Attributes(m map[string]interface{}, attrs ...string) string {
	for k, v := range m {
		attrs = append(attrs, fmt.Sprintf("%s=%s", k, Quote(v)))
	}

	if len(attrs) == 0 {
		return ""
	}

	return "[" + strings.Join(attrs, ",") + "]"
}

// ImageMapFn
// id is the ID of the node the image will be placed
// imageType is the type of image you will be placing
type ImageMapFn func(imageType ImageType) Image

func ImageMapper(fn ImageMapFn, imageType ImageType) string {
	image := fn(imageType)
	legendCh <- image
	return image.Path
}

func OrderLegend(in <-chan Image) <-chan Image {
	out := make(chan Image)
	go func() {
		defer close(out)

		// Collect all legends from the input channel
		var legends []Image
		for legend := range in {
			legends = append(legends, legend)
		}

		// Sort legends alphabetically by Label
		sort.Slice(legends, func(i, j int) bool {
			return legends[i].Label < legends[j].Label
		})

		// Deduplicate legends
		var seen = make(map[string]bool)
		for _, legend := range legends {
			if !seen[legend.Label] {
				seen[legend.Label] = true
				out <- legend
			}
		}
	}()
	return out
}

// Writes the Graphviz dot representation of the graph to the specified file
func (t Graph) WriteDotFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	return t.WriteDot(file)
}

func (t Graph) WriteDot(writer io.Writer) error {
	if _, err := io.WriteString(writer, t.Dot()); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	return nil
}

// Generates an SVG from the dot file using the dot binary
func (t Graph) WriteSVG(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	return t.WriteImage(file, "svg")
}

// Generates an SVG from the dot file using the dot binary
func (t Graph) WritePNG(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	return t.WriteImage(file, "png")
}

func (t Graph) WriteImage(writer io.Writer, typ string) error {
	cmd := exec.Command("dot", "-T"+typ)
	cmd.Stdin = strings.NewReader(t.Dot())

	out := &bytes.Buffer{}
	cmd.Stdout = out

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running dot command: %v", err)
	}

	_, err = io.Copy(writer, out)
	if err != nil {
		return fmt.Errorf("error copying output: %v", err)
	}

	return nil
}
