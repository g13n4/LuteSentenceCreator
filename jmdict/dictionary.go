package jmdict

import (
	"strings"
)

const DictionaryNews = "news"
const DictionaryIchi = "ichi"
const DictionarySpec = "spec"
const DictionaryGai = "gai"
const DictionaryNF = "nf"

type Dictionary struct {
	Id          int
	Category    string
	Description string
}

var emptyDictionary = Dictionary{
	Id:          9,
	Category:    "",
	Description: "",
}

type DictionaryPool struct {
	dictionaries map[string]*Dictionary
	categories   []string
}

func NewDictionaryPool() *DictionaryPool {
	dp := DictionaryPool{
		dictionaries: make(map[string]*Dictionary),
		categories:   make([]string, 0),
	}
	dp.FillPool()
	return &dp
}

func (dp *DictionaryPool) GetCategory(name string) string {
	for _, category := range dp.categories {
		if strings.HasPrefix(name, category) {
			return category
		}
	}
	return ""
}

func (dp *DictionaryPool) GetDictionary(name string) *Dictionary {
	cat := dp.GetCategory(name)
	dict, ok := dp.dictionaries[cat]
	if ok {
		return dict
	}

	return &emptyDictionary
}

func (dp *DictionaryPool) GetAllDictionaries() *[]Dictionary {
	values := make([]Dictionary, 0, len(dp.dictionaries))
	for _, v := range dp.dictionaries {
		values = append(values, *v)
	}
	return &values
}

func (dp *DictionaryPool) FillPool() {
	dp.categories = []string{DictionaryNews, DictionaryIchi, DictionarySpec, DictionaryGai, DictionaryNF}
	dp.dictionaries[DictionaryNews] = &Dictionary{
		Id:          1,
		Category:    DictionaryNews,
		Description: "appears in the \"wordfreq\" file compiled by Alexandre Girardi from the Mainichi Shimbun. (See the Monash ftp archive for a copy.) Words in the first 12,000 in that file are marked \"news1\" and words in the second 12,000 are marked \"news2\".",
	}

	dp.dictionaries[DictionaryIchi] = &Dictionary{
		Id:          2,
		Category:    DictionaryIchi,
		Description: "appears in the \"Ichimango goi bunruishuu\", Senmon Kyouiku Publishing, Tokyo, 1998.  (The entries marked \"ichi2\" were demoted from ichi1 because they were observed to have low frequencies in the WWW and newspapers.)",
	}

	dp.dictionaries[DictionarySpec] = &Dictionary{
		Id:          3,
		Category:    DictionarySpec,
		Description: "a small number of words use this marker when they are detected as being common, but are not included in other lists.",
	}

	dp.dictionaries[DictionaryGai] = &Dictionary{
		Id:          4,
		Category:    DictionaryGai,
		Description: "common loanwords, based on the wordfreq file.",
	}

	dp.dictionaries[DictionaryNF] = &Dictionary{
		Id:          5,
		Category:    DictionaryNF,
		Description: "this is an indicator of frequency-of-use ranking in the wordfreq file. \"xx\" is the number of the set of 500 words in which the entry can be found, with \"01\" assigned to the first 500, \"02\" to the second, and so on. (The entries with news1, ichi1, spec1, spec2 and gai1 values are marked with a \"(P)\" in the EDICT and EDICT2 files.)",
	}

	dp.dictionaries["&missing"] = &emptyDictionary
}
