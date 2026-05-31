package jmdict

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const DictionaryNews = "news"
const DictionaryIchi = "ichi"
const DictionarySpec = "spec"
const DictionaryGai = "gai"
const DictionaryNF = "nf"

const DictionaryNewsValue = 1
const DictionaryIchiValue = 2
const DictionarySpecValue = 3
const DictionaryGaiValue = 4
const DictionaryNFValue = 5

type Dictionary struct {
	Id       int
	Name     string
	Category int
	Number   int
}

func (d *Dictionary) String() string {
	return fmt.Sprintf("Id: %d, Name: %s, Category: %d, Number: %d", d.Id, d.Name, d.Category, d.Number)
}

type DictionaryCategory struct {
	Id          int
	Name        string
	Description string
}

func (d *DictionaryCategory) String() string {
	return fmt.Sprintf("Id: %d, Name: %s", d.Id, d.Name)
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
		dictionaries:  make(map[string]*Dictionary),
		categories:    make(map[string]*DictionaryCategory),
		categoryNames: make([]string, 0),
	}
	dp.FillPool()
	return &dp
}

func (dp *DictionaryPool) GetCategory(name string) *DictionaryCategory {
	for _, category := range dp.categoryNames {
		if strings.HasPrefix(name, category) {
			return dp.categories[category]
		}
	}
	return dp.categories["&missing"]
}

func (dp *DictionaryPool) GetCategoryName(name string) string {
	for _, category := range dp.categoryNames {
		if strings.HasPrefix(name, category) {
			return category
		}
	}
	return ""
}

func (dp *DictionaryPool) GetDictionaryId(dictId, dictCatId int) int {
	dictionaryOffset := dictCatId * 100
	return dictionaryOffset + dictId
}

func (dp *DictionaryPool) getDictionary(name string, categoryId int) *Dictionary {
	match := regexp.MustCompile(`\D+(\d+)`)
	numbers := match.FindAllStringSubmatch(name, 1)
	dictionaryValue, err := strconv.Atoi(numbers[0][1])
	if err != nil {
		return &Dictionary{
			Id:       99,
			Name:     name,
			Category: emptyDictionary.Id,
			Number:   0,
		}
	}

	return &Dictionary{
		Id:       dp.GetDictionaryId(dictionaryValue, categoryId),
		Name:     name,
		Category: categoryId,
		Number:   dictionaryValue,
	}
}

func (dp *DictionaryPool) GetDictionaryData(name string) (*Dictionary, *DictionaryCategory) {
	cat := dp.GetCategoryName(name)
	dictCat, ok := dp.categories[cat]
	if !ok {
		dictCat = &emptyDictionary
	}
	dict, ok := dp.dictionaries[name]
	if ok {
		return dict, dictCat
	}
	dict = dp.getDictionary(name, dictCat.Id)
	dp.dictionaries[name] = dict

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
		Id:          DictionaryNewsValue,
		Name:        DictionaryNews,
		Description: "appears in the \"wordfreq\" file compiled by Alexandre Girardi from the Mainichi Shimbun. (See the Monash ftp archive for a copy.) Words in the first 12,000 in that file are marked \"news1\" and words in the second 12,000 are marked \"news2\".",
	}

	dp.categories[DictionaryIchi] = &DictionaryCategory{
		Id:          DictionaryIchiValue,
		Name:        DictionaryIchi,
		Description: "appears in the \"Ichimango goi bunruishuu\", Senmon Kyouiku Publishing, Tokyo, 1998.  (The entries marked \"ichi2\" were demoted from ichi1 because they were observed to have low frequencies in the WWW and newspapers.)",
	}

	dp.categories[DictionarySpec] = &DictionaryCategory{
		Id:          DictionarySpecValue,
		Name:        DictionarySpec,
		Description: "a small number of words use this marker when they are detected as being common, but are not included in other lists.",
	}

	dp.categories[DictionaryGai] = &DictionaryCategory{
		Id:          DictionaryGaiValue,
		Name:        DictionaryGai,
		Description: "common loanwords, based on the wordfreq file.",
	}

	dp.categories[DictionaryNF] = &DictionaryCategory{
		Id:          DictionaryNFValue,
		Name:        DictionaryNF,
		Description: "this is an indicator of frequency-of-use ranking in the wordfreq file. \"xx\" is the number of the set of 500 words in which the entry can be found, with \"01\" assigned to the first 500, \"02\" to the second, and so on. (The entries with news1, ichi1, spec1, spec2 and gai1 values are marked with a \"(P)\" in the EDICT and EDICT2 files.)",
	}

	dp.categories["&missing"] = &emptyDictionary
}
