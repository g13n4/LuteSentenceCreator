package utils

import (
	"fmt"
	"strings"
)

const Null = "NULL"

type MyIntPtr *int
type PostgresID int

type Stringer interface {
	String() string
}

func FormatStringFromArray(title string, list []string) string {
	var output string
	if len(list) != 0 {
		output = strings.Join(list, "; ")
		output = fmt.Sprintf("%s: %s\n", title, output)
	}
	return output
}

func FormatIntNullIfNil(v *int) string {
	if v == nil {
		return Null
	}
	return fmt.Sprintf("%v", *v)
}

func GetUTFValue(v string) int {
	val := []rune(v)
	return int(val[0])
}
