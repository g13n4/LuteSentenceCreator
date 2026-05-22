package jmdict

import (
	"regexp"
	"strconv"
	"strings"
)

const DictionaryNews = "news"
const DictionaryIchi = "ichi"
const DictionarySpec = "spec"
const DictionaryGai = "gai"
const DictionaryNF = "nf"

type Dictionary struct {
	Id   int
	Name string
}

type DictionaryCategory struct {
	Id          int
	Name        string
	Description string
}

var emptyDictionary = DictionaryCategory{
	Id:          999,
	Name:        "",
	Description: "",
}

type DictionaryPool struct {
	dictionaries  map[string]*Dictionary
	categories    map[string]*DictionaryCategory
	categoryNames []string
}

func NewDictionaryPool() *DictionaryPool {
	dp := DictionaryPool{
		categories:    make(map[string]*DictionaryCategory),
		categoryNames: make([]string, 0),
	}
	dp.FillPool()
	return &dp
}

func (dp *DictionaryPool) GetCategory(name string) string {
	for _, category := range dp.categoryNames {
		if strings.HasPrefix(name, category) {
			return category
		}
	}
	return ""
}

func (dp *DictionaryPool) getDictionary(name string, categoryId int) *Dictionary {
	categoryId = categoryId * 100
	match := regexp.MustCompile("\\w+(\\d+)")
	numbers := match.FindAllString(name, 1)
	dictionaryValue, err := strconv.Atoi(numbers[0])
	if err != nil {
		return &Dictionary{
			Id:   99,
			Name: name,
		}
	}

	return &Dictionary{
		Id:   categoryId + dictionaryValue,
		Name: name,
	}
}

func (dp *DictionaryPool) GetDictionaryData(name string) (*Dictionary, *DictionaryCategory) {
	cat := dp.GetCategory(name)
	dictCat, ok := dp.categories[cat]
	if !ok {
		dictCat = &emptyDictionary
	}
	dict, ok := dp.dictionaries[name]
	if ok {
		return dict, dictCat
	}
	dict = dp.getDictionary(name, dictCat.Id)

	return dict, dictCat
}

func (dp *DictionaryPool) GetAllCategories() *[]*DictionaryCategory {
	values := make([]*DictionaryCategory, 0)
	for _, v := range dp.categories {
		values = append(values, v)
	}
	return &values
}

func (dp *DictionaryPool) GetAllDictionaries() *[]*Dictionary {
	values := make([]*Dictionary, 0)
	for _, v := range dp.dictionaries {
		values = append(values, v)
	}
	return &values
}

func (dp *DictionaryPool) FillPool() {
	dp.categoryNames = []string{DictionaryNews, DictionaryIchi, DictionarySpec, DictionaryGai, DictionaryNF}
	dp.categories[DictionaryNews] = &DictionaryCategory{
		Id:          1,
		Name:        DictionaryNews,
		Description: "appears in the \"wordfreq\" file compiled by Alexandre Girardi from the Mainichi Shimbun. (See the Monash ftp archive for a copy.) Words in the first 12,000 in that file are marked \"news1\" and words in the second 12,000 are marked \"news2\".",
	}

	dp.categories[DictionaryIchi] = &DictionaryCategory{
		Id:          2,
		Name:        DictionaryIchi,
		Description: "appears in the \"Ichimango goi bunruishuu\", Senmon Kyouiku Publishing, Tokyo, 1998.  (The entries marked \"ichi2\" were demoted from ichi1 because they were observed to have low frequencies in the WWW and newspapers.)",
	}

	dp.categories[DictionarySpec] = &DictionaryCategory{
		Id:          3,
		Name:        DictionarySpec,
		Description: "a small number of words use this marker when they are detected as being common, but are not included in other lists.",
	}

	dp.categories[DictionaryGai] = &DictionaryCategory{
		Id:          4,
		Name:        DictionaryGai,
		Description: "common loanwords, based on the wordfreq file.",
	}

	dp.categories[DictionaryNF] = &DictionaryCategory{
		Id:          5,
		Name:        DictionaryNF,
		Description: "this is an indicator of frequency-of-use ranking in the wordfreq file. \"xx\" is the number of the set of 500 words in which the entry can be found, with \"01\" assigned to the first 500, \"02\" to the second, and so on. (The entries with news1, ichi1, spec1, spec2 and gai1 values are marked with a \"(P)\" in the EDICT and EDICT2 files.)",
	}

	dp.categories["&missing"] = &emptyDictionary
}
