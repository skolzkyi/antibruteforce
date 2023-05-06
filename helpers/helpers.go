package helpers

import (
	"strings"
)

func StringBuild(input ...string) string {
	var sb strings.Builder

	for _, str := range input {
		sb.WriteString(str)
	}

	return sb.String()
}
