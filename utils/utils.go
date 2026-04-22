package utils

import (
	"fmt"
	"strings"
)

func FormatStringFromArray(title string, list []string) string {
	var output string
	if len(list) != 0 {
		output = strings.Join(list, "; ")
		output = fmt.Sprintf("%s: %s\n", title, output)
	}
	return output
}
