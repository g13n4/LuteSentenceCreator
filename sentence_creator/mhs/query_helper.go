package mhs

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v5"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

func NewQueryHelper(c *echo.Context) (*QueryHelper, error) {
	var qh QueryHelper
	qh.JLPT = c.QueryParam("jlpt")
	qh.Frequency = c.QueryParam("freq")
	qh.Grade = c.QueryParam("grade")
	qh.StrokeCount = c.QueryParam("stroke_count")
	qh.Dictionary = c.QueryParam("dictionary")
	qh.DictionaryCategory = c.QueryParam("category")
	if qh.IsEmpty() {
		return nil, errors.New("all query parameters are empty")
	}
	return &qh, nil
}

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

	return o
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

func (q *QueryHelper) getSQLCondition() string {
	var tableName string
	if q.IsKanji() {
		tableName = "k" + "."
	} else {
		tableName = "dme" + "."
	}

	if q.JLPT != "" {
		if q.JLPT == "0" {
			return fmt.Sprintf("%sjlpt IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sjlpt = %v ", tableName, q.JLPT)
	}

	if q.Frequency != "" {
		if q.Frequency == "0" {
			return fmt.Sprintf("%sfreq IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sfreq = %v ", tableName, q.Frequency)
	}

	if q.Grade != "" {
		if q.Grade == "0" {
			return fmt.Sprintf("%sgrade IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sgrade = %v ", tableName, q.Grade)
	}

	if q.StrokeCount != "" {
		if q.StrokeCount == "0" {
			return fmt.Sprintf("%sstroke_count IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sstroke_count = %v ", tableName, q.StrokeCount)
	}

	if q.DictionaryCategory != "" {
		return fmt.Sprintf("%sdc_id = %v ", tableName, q.DictionaryCategory)
	}

	return fmt.Sprintf("%sd_id = %v ", tableName, q.Dictionary)
}

func (q *QueryHelper) IsKanji() bool {
	if q.Dictionary != "" || q.DictionaryCategory != "" {
		return false
	}
	return true
}

func (q *QueryHelper) IsEmpty() bool {
	if q.JLPT == "" &&
		q.Frequency == "" &&
		q.Grade == "" &&
		q.StrokeCount == "" &&
		q.Dictionary == "" &&
		q.DictionaryCategory == "" {
		return true
	}
	return false
}

func (q *QueryHelper) CreateQuery() string {
	if q.IsKanji() {
		return fmt.Sprintf("SELECT DISTINCT smr.r_id, smr.s_id from sentences__mtm__readings smr JOIN readings__mtm__kanjis rmk ON rmk.r_id = smr.r_id JOIN kanjis k ON rmk.k_id = k.id WHERE %s", q.getSQLCondition())
	}

	return fmt.Sprintf("SELECT DISTINCT smr.r_id, smr.s_id from sentences__mtm__readings smr JOIN readings r ON smr.r_id = r.id JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry WHERE %s", q.getSQLCondition())
}

func (q *QueryHelper) GetPreallocSize() int {
	if q.IsKanji() {
		return 3000
	}

	return 50000
}
