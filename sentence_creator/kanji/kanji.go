package kanji

import (
	"fmt"
	"strings"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

const KanjiNodeName = "character"

type Kanji struct {
	Literal string `xml:"literal"`

	JLPT        *int `xml:"misc>jlpt"`
	Frequency   *int `xml:"misc>freq"`
	Grade       *int `xml:"misc>grade"`
	StrokeCount *int `xml:"misc>stroke_count"`
}

func (k *Kanji) IsPopular() bool {
	if k.JLPT != nil || k.Frequency != nil || k.Grade != nil {
		return true
	}
	return false
}

func (k *Kanji) String() string {
	var info []string

	if k.JLPT != nil {
		info = append(info, fmt.Sprintf("JLPT: %v", *k.JLPT))
	}

	if k.Frequency != nil {
		info = append(info, fmt.Sprintf("Frequency: %v", *k.Frequency))
	}

	if k.Grade != nil {
		info = append(info, fmt.Sprintf("Grade: %v", *k.Grade))
	}

	if k.StrokeCount != nil {
		info = append(info, fmt.Sprintf("Stroke Count: %v", *k.StrokeCount))
	}

	infoStr := strings.Join(info, ", ")

	if infoStr != "" {
		return fmt.Sprintf("%s: %s", k.Literal, infoStr)
	}

	return k.Literal
}

func (k *Kanji) ToSaveSQL() string {
	var info []string

	info = append(info, k.Literal)

	info = append(info, utils.FormatIntNullIfNil(k.JLPT))
	info = append(info, utils.FormatIntNullIfNil(k.Frequency))
	info = append(info, utils.FormatIntNullIfNil(k.Grade))
	info = append(info, utils.FormatIntNullIfNil(k.StrokeCount))

	inner := strings.Join(info, ", ")

	return fmt.Sprintf("(%s)", inner)
}
