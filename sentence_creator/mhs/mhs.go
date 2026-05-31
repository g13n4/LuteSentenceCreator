package mhs

import (
	"fmt"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

type QueryHelper struct {
	JLPT        string
	Frequency   string
	Grade       string
	StrokeCount string

	Dictionary         string
	DictionaryCategory string
}

func (q *QueryHelper) String() string {
	var o string

	if q.JLPT != "" {
		o += fmt.Sprintf("jlpt-%s", q.JLPT)
	}

	if q.Frequency != "" {
		o += fmt.Sprintf("frequency-%s", q.Frequency)
	}

	if q.Grade != "" {
		o += fmt.Sprintf("grade-%s", q.Grade)
	}

	if q.StrokeCount != "" {
		o += fmt.Sprintf("strokes-%s", q.StrokeCount)
	}

	if q.Dictionary != "" {
		o += fmt.Sprintf("dictionary-%s", q.Dictionary)
	}

	if q.DictionaryCategory != "" {
		o += fmt.Sprintf("dictionary_category-%s", q.DictionaryCategory)
	}

	return ""
}

func (q *QueryHelper) StringFull() string {
	return fmt.Sprintf(
		"JLPT: %v; Frequency: %v; Grade: %v; StrokeCount: %v; Dictionary: %v; Dictionary Category: %v",
		utils.StringOrDash(q.JLPT),
		utils.StringOrDash(q.Frequency),
		utils.StringOrDash(q.Grade),
		utils.StringOrDash(q.StrokeCount),
		utils.StringOrDash(q.Dictionary),
		utils.StringOrDash(q.DictionaryCategory),
	)
}

func (q *QueryHelper) GetSQLCondition(tableName string) string {
	if tableName != "" {
		tableName = "." + tableName
	}

	if q.JLPT != "" {
		if q.JLPT == "0" {
			return fmt.Sprintf("%s.jlpt IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%s.jlpt = %v ", tableName, q.JLPT)
	}

	if q.Frequency != "" {
		if q.Frequency == "0" {
			return fmt.Sprintf("%s.freq IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%s.freq = %v ", tableName, q.Frequency)
	}

	if q.Grade != "" {
		if q.Grade == "0" {
			return fmt.Sprintf("%s.grade IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%s.grade = %v ", tableName, q.Grade)
	}

	if q.StrokeCount != "" {
		if q.StrokeCount == "0" {
			return fmt.Sprintf("%s.stroke_count IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%s.stroke_count = %v ", tableName, q.StrokeCount)
	}

	if q.DictionaryCategory != "" {
		return fmt.Sprintf("%s.dc_id = %v ", tableName, q.DictionaryCategory)
	}

	return fmt.Sprintf("%s.d_id = %v ", tableName, q.Dictionary)
}

func (q *QueryHelper) IsKanji() bool {
	if q.Dictionary != "" || q.DictionaryCategory != "" {
		return false
	}
	return true
}

func (q *QueryHelper) Clean() {
	q.JLPT = ""
	q.Frequency = ""
	q.Grade = ""
	q.StrokeCount = ""
	q.Dictionary = ""
	q.DictionaryCategory = ""
}
