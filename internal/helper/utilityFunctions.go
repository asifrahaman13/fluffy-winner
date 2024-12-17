package helper

import (
	"bytes"
	"strings"
)

func IsSentenceEnd(buffer bytes.Buffer) bool {
	return strings.HasSuffix(buffer.String(), ".") || strings.HasSuffix(buffer.String(), "!") || strings.HasSuffix(buffer.String(), "?")
}
