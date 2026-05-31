package mhs

import (
	"fmt"

	"github.com/g13n4/LuteSentencePicker/utils"
)

type QueryHelper struct {
	JLPT        *int
	Frequency   *int
	Grade       *int
	StrokeCount *int

	Dictionary         *int
	DictionaryCategory *int
}

func (q *QueryHelper) String() string {
	return fmt.Sprintf(
		"JLPT: %v; Frequency: %v; Grade: %v; StrokeCount: %v; Dictionary: %v; Dictionary Category: %v",
		utils.PIntegerToSafeString(q.JLPT),
		utils.PIntegerToSafeString(q.Frequency),
		utils.PIntegerToSafeString(q.Grade),
		utils.PIntegerToSafeString(q.StrokeCount),
		utils.PIntegerToSafeString(q.Dictionary),
		utils.PIntegerToSafeString(q.DictionaryCategory),
	)
}

func (q *QueryHelper) GetSQLCondition(tableName string) string {
	if tableName != "" {
		tableName = "." + tableName
	}

	if q.JLPT != nil {
		return fmt.Sprintf("%s.jlpt = %v ", tableName, *q.JLPT)
	}

	if q.Frequency != nil {
		return fmt.Sprintf("%s.freq = %v ", tableName, *q.Frequency)
	}

	if q.Grade != nil {
		return fmt.Sprintf("%s.grade = %v ", tableName, *q.Grade)
	}

	if q.StrokeCount != nil {
		return fmt.Sprintf("%s.stroke_count = %v ", tableName, *q.StrokeCount)
	}

	if q.DictionaryCategory != nil {
		return fmt.Sprintf("%s.dc_id = %v ", tableName, *q.DictionaryCategory)
	}

	return fmt.Sprintf("%s.d_id = %v ", tableName, *q.Dictionary)
}

func (q *QueryHelper) IsKanji() bool {
	if q.Dictionary != nil || q.DictionaryCategory != nil {
		return false
	}
	return true
}
