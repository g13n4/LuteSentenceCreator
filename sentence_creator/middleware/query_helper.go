package middleware

import (
	"errors"
	"fmt"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

type ParamQueryExtractor interface {
	QueryParam(name string) string
}

func NewQueryHelper(c ParamQueryExtractor) (*QueryHelper, error) {
	var qh QueryHelper

	qh.MaxSentencesInJoin = utils.GetEnvIntValue("MHS_MAX_SENTENCES_PER_VALUE", 30)

	qh.UseMHS = c.QueryParam("mhs") == "true"

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

	MaxSentencesInJoin int

	UseMHS bool
}

func (qh *QueryHelper) String() string {
	var o string

	if qh.JLPT != "" {
		o += fmt.Sprintf("jlpt-%s", qh.JLPT)
	}

	if qh.Frequency != "" {
		o += fmt.Sprintf("frequency-%s", qh.Frequency)
	}

	if qh.Grade != "" {
		o += fmt.Sprintf("grade-%s", qh.Grade)
	}

	if qh.StrokeCount != "" {
		o += fmt.Sprintf("strokes-%s", qh.StrokeCount)
	}

	if qh.Dictionary != "" {
		o += fmt.Sprintf("dictionary-%s", qh.Dictionary)
	}

	if qh.DictionaryCategory != "" {
		o += fmt.Sprintf("dictionary_category-%s", qh.DictionaryCategory)
	}

	return o
}

func (qh *QueryHelper) StringFull() string {
	return fmt.Sprintf(
		"JLPT: %v; Frequency: %v; Grade: %v; StrokeCount: %v; Dictionary: %v; Dictionary Category: %v",
		utils.StringOrDash(qh.JLPT),
		utils.StringOrDash(qh.Frequency),
		utils.StringOrDash(qh.Grade),
		utils.StringOrDash(qh.StrokeCount),
		utils.StringOrDash(qh.Dictionary),
		utils.StringOrDash(qh.DictionaryCategory),
	)
}

func (qh *QueryHelper) getSQLCondition() string {
	var tableName string
	if qh.IsKanji() {
		tableName = "k" + "."
	} else {
		tableName = "dme" + "."
	}

	if qh.JLPT != "" {
		if qh.JLPT == "0" {
			return fmt.Sprintf("%sjlpt IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sjlpt = %v ", tableName, qh.JLPT)
	}

	if qh.Frequency != "" {
		if qh.Frequency == "0" {
			return fmt.Sprintf("%sfreq IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sfreq = %v ", tableName, qh.Frequency)
	}

	if qh.Grade != "" {
		if qh.Grade == "0" {
			return fmt.Sprintf("%sgrade IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sgrade = %v ", tableName, qh.Grade)
	}

	if qh.StrokeCount != "" {
		if qh.StrokeCount == "0" {
			return fmt.Sprintf("%sstroke_count IS NOT NULL", tableName)
		}

		return fmt.Sprintf("%sstroke_count = %v ", tableName, qh.StrokeCount)
	}

	if qh.DictionaryCategory != "" {
		return fmt.Sprintf("%sdc_id = %v ", tableName, qh.DictionaryCategory)
	}

	return fmt.Sprintf("%sd_id = %v ", tableName, qh.Dictionary)
}

func (qh *QueryHelper) IsKanji() bool {
	if qh.Dictionary != "" || qh.DictionaryCategory != "" {
		return false
	}
	return true
}

func (qh *QueryHelper) IsEmpty() bool {
	if qh.JLPT == "" &&
		qh.Frequency == "" &&
		qh.Grade == "" &&
		qh.StrokeCount == "" &&
		qh.Dictionary == "" &&
		qh.DictionaryCategory == "" {
		return true
	}
	return false
}

func (qh *QueryHelper) addWhereClause(sql string) string {
	return sql + " WHERE s.isFiltered = false " + qh.getSQLCondition()
}

func (qh *QueryHelper) addLimitClause(sql string, limit int) string {
	if limit == 0 {
		return sql
	}
	return sql + fmt.Sprintf(" LIMIT %v", limit)
}

func (qh *QueryHelper) CreateMHSQuery() string {
	var sql string
	if qh.IsKanji() {
		sql = fmt.Sprintf("SELECT DISTINCT smr.r_id, smr.s_id from kanjis k JOIN readings__mtm__kanjis rmk ON rmk.k_id = k.id INNER JOIN LATERAL (SELECT DISTINCT smr.r_id, smr.s_id FROM sentences__mtm__readings smr WHERE rmk.r_id = smr.r_id ORDER BY smr.r_id, smr.s_id LIMIT %v ) smr ON rmk.r_id = smr.r_id ORDER BY smr.r_id, smr.s_id", qh.MaxSentencesInJoin)
	} else {
		sql = fmt.Sprintf("SELECT DISTINCT smr.r_id, smr.s_id from readings r JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry INNER JOIN LATERAL ( SELECT DISTINCT smr.r_id, smr.s_id FROM sentences__mtm__readings smr WHERE r.id = smr.r_id ORDER BY smr.r_id, smr.s_id LIMIT %v ) smr ON r.id = smr.r_id", qh.MaxSentencesInJoin)
	}

	sql = qh.addWhereClause(sql)

	return sql + " ORDER BY smr.r_id, smr.s_id"
}

func (qh *QueryHelper) CreateSimpleQuery(limit int) string {
	var sql string
	if qh.IsKanji() {
		sql = fmt.Sprintf("SELECT DISTINCT sentence from kanjis k JOIN readings__mtm__kanjis rmk ON rmk.k_id = k.id INNER JOIN LATERAL (SELECT DISTINCT smr.r_id, smr.s_id FROM sentences__mtm__readings smr WHERE rmk.r_id = smr.r_id ORDER BY smr.r_id, smr.s_id LIMIT %v ) smr ON rmk.r_id = smr.r_id JOIN sentences s ON smr.s_id = s.id", qh.MaxSentencesInJoin)
	} else {
		sql = fmt.Sprintf("SELECT DISTINCT sentence from readings r JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry INNER JOIN LATERAL ( SELECT DISTINCT smr.r_id, smr.s_id FROM sentences__mtm__readings smr WHERE r.id = smr.r_id ORDER BY smr.r_id, smr.s_id LIMIT %v ) smr ON r.id = smr.r_id JOIN sentences s ON smr.s_id = s.id", qh.MaxSentencesInJoin)
	}

	sql = qh.addWhereClause(sql)
	sql = qh.addLimitClause(sql, limit)

	return sql
}
