package utf8

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// Cleans malformed utf8 data (typically from python ewwww) that has doubly
// encoded utf8/unicode characters
// Example: "This is Stephen\\u2019s"
// Output: "This is Stephen’s"
func CleanDoubleEscapedUtf8(source string) (string, error) {
	out := &bytes.Buffer{}
	var rr rune
	index := -1
	for i, r := range source {
		if index > -1 && i <= index+5 {
			if r >= 48 && r <= 57 {
				rr = rr*16 + (r - 48)
			}

			if r >= 97 && r <= 102 {
				rr = rr*16 + (r - 97)
			}

			if i == index+5 {
				if !utf8.ValidRune(rr) {
					return "", fmt.Errorf("Not valid utf8 character '%s'", source[index:index+6])
				}
				out.WriteRune(rr)
				rr = 0
				index = -1
			}
			continue
		}

		if i+1 < len(source) && source[i:i+2] == "\\u" {
			index = i
			continue
		}

		out.WriteRune(r)
	}

	return out.String(), nil
}
