package jmdict

import (
	"fmt"
	"strings"

	"github.com/g13n4/LuteSentenceCreator/utils"
)

const EntryNodeName = "entry"

func CreateDictionaryEntry(name string) *DictionaryEntry {
	var category string

	category = "news"
	if strings.HasPrefix(name, category) {
		return &DictionaryEntry{
			Name:        name,
			Category:    category,
			Description: "appears in the \"wordfreq\" file compiled by Alexandre Girardi from the Mainichi Shimbun. (See the Monash ftp archive for a copy.) Words in the first 12,000 in that file are marked \"news1\" and words in the second 12,000 are marked \"news2\".",
		}
	}

	category = "ichi"
	if strings.HasPrefix(name, category) {
		return &DictionaryEntry{
			Name:        name,
			Category:    category,
			Description: "appears in the \"Ichimango goi bunruishuu\", Senmon Kyouiku Publishing, Tokyo, 1998.  (The entries marked \"ichi2\" were demoted from ichi1 because they were observed to have low frequencies in the WWW and newspapers.)",
		}
	}

	category = "spec"
	if strings.HasPrefix(name, category) {
		return &DictionaryEntry{
			Name:        name,
			Category:    category,
			Description: "a small number of words use this marker when they are detected as being common, but are not included in other lists.",
		}
	}

	category = "gai"
	if strings.HasPrefix(name, category) {
		return &DictionaryEntry{
			Name:        name,
			Category:    category,
			Description: "common loanwords, based on the wordfreq file.",
		}
	}

	category = "nf"
	if strings.HasPrefix(name, category) {
		return &DictionaryEntry{
			Name:        name,
			Category:    category,
			Description: "this is an indicator of frequency-of-use ranking in the wordfreq file. \"xx\" is the number of the set of 500 words in which the entry can be found, with \"01\" assigned to the first 500, \"02\" to the second, and so on. (The entries with news1, ichi1, spec1, spec2 and gai1 values are marked with a \"(P)\" in the EDICT and EDICT2 files.)",
		}
	}

	return &DictionaryEntry{
		Name: name,
	}

}

type DictionaryEntry struct {
	Name        string
	Category    string
	Description string
}

func (de DictionaryEntry) ToSQL() string {
	return fmt.Sprintf("(%s, %s, %s)", de.Name, de.Category, de.Description)
}

type Entry struct {
	EntryId      int      `xml:"ent_seq"`
	KanjiReading []string `xml:"k_ele>keb"`
	Reading      []string `xml:"r_ele>reb"`

	InDictionaryKanji   []string `xml:"k_ele>ke_pri"`
	InDictionaryReading []string `xml:"r_ele>re_pri"`
}

func (e *Entry) String() string {
	krInfo := utils.FormatStringFromArray("Reading [incl. kanji]", e.KanjiReading)
	rInfo := utils.FormatStringFromArray("Reading [w/o kanji]", e.Reading)
	dkInfo := utils.FormatStringFromArray("In Dictionary [incl. kanji]", e.InDictionaryKanji)
	drInfo := utils.FormatStringFromArray("In Dictionary [w/o kanji]", e.InDictionaryReading)

	return fmt.Sprintf(
		"Entry id: %d\n%s%s%s%s",
		e.EntryId,
		krInfo,
		rInfo,
		dkInfo,
		drInfo,
	)
}

func (e *Entry) IsPopular() bool {
	if len(e.InDictionaryKanji) != 0 || len(e.InDictionaryReading) != 0 {
		return true
	}
	return false
}

func (e *Entry) ToSQL() string {
	return ""
}
