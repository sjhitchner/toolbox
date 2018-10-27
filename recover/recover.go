package recover

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
)

// Recover stack trace from the stack
func RecoverStackTrace(panicReason interface{}) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Recover from panic: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	return buffer.String()
}

func RecoverError(panicReason interface{}) error {
	switch t := panicReason.(type) {
	case string:
		return errors.New(t)
	case error:
		return t
	default:
		return errors.New("Unknown error")
	}
}
