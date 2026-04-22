package jmdict

import (
	"fmt"

	"github.com/g13n4/LuteSentenceCreator/utils"
)

const EntryNodeName = "entry"

type Entry struct {
	EntryId      int      `xml:"ent_seq"`
	KanjiReading []string `xml:"k_ele>keb"`
	Reading      []string `xml:"r_ele>reb"`

	InDictionaryKanji   []string `xml:"k_ele>ke_pri"`
	InDictionaryReading []string `xml:"r_ele>re_pri"`
}

func (e Entry) String() string {
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

func (e Entry) IsPopular() bool {
	if len(e.InDictionaryKanji) != 0 || len(e.InDictionaryReading) != 0 {
		return true
	}
	return false
}
