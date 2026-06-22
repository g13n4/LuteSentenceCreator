package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const Null = "NULL"

type MyIntPtr *int
type PostgresID int

type Stringer interface {
	String() string
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

func IntegerToSafeString(v int) string {
	return fmt.Sprintf("%v", v)
}

func StringOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func GetEnvIntValue(key string, defValue int) int {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("%s environment variable is not set. Using default", key)
	} else {
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("%s environment variable is not correct. Using default", key)
		} else {
			defValue = val
		}
	}
	return defValue
}
