package kanji

import (
	"fmt"
	"strings"
)

const KanjiNodeName = "character"

type Kanji struct {
	Literal string `xml:"literal"`

	JLPT      *int `xml:"misc>jlpt"`
	Frequency *int `xml:"misc>freq"`
	Grade     *int `xml:"misc>grade"`
}

func (k Kanji) IsPopular() bool {
	if k.JLPT != nil || k.Frequency != nil || k.Grade != nil {
		return true
	}
	return false
}

func (k Kanji) String() string {
	var info []string

	if k.JLPT != nil {
		info = append(info, fmt.Sprintf("JLPT: %d", *k.JLPT))
	}

	if k.Frequency != nil {
		info = append(info, fmt.Sprintf("Frequency: %d", *k.Frequency))
	}

	if k.Grade != nil {
		info = append(info, fmt.Sprintf("Grade: %d", *k.Grade))
	}
	infoStr := strings.Join(info, ", ")

	if infoStr != "" {
		return fmt.Sprintf("%s: %s", k.Literal, infoStr)
	}

	return k.Literal
}
