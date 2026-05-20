package jmdict

import (
	"encoding/xml"
	"fmt"
	"maps"
	"slices"
	"strings"
)

const EntryNodeName = "entry"

type WordEntry interface {
	IsInDictionary() bool
	NewsInDictionary() bool
	String() string
}

func NewsInDictionary(dictionaries *[]string) bool {
	for _, word := range *dictionaries {
		if strings.Contains(word, DictionaryNews) {
			return true
		}
	}
	return false
}

type Reading struct {
	Word       string
	Dictionary []string
	IsKanji    bool
	CombinedId int
}

func (r *Reading) IsInDictionary() bool {
	return len(r.Dictionary) == 0
}

func (r *Reading) NewsInDictionary() bool {
	return NewsInDictionary(&r.Dictionary)
}

func (r *Reading) String() string {
	if len(r.Dictionary) != 0 {
		return fmt.Sprintf(
			"%s [%s]",
			r.Word,
			strings.Join(r.Dictionary, ", "),
		)
	}
	return r.Word
}

type Entry struct {
	EntryId  int       `xml:"ent_seq"`
	Readings []Reading `xml:"-"`
}

type RElement struct {
	Reb   string   `xml:"reb"`
	RePri []string `xml:"re_pri"`
}

type KElement struct {
	Keb   string   `xml:"keb"`
	KePri []string `xml:"ke_pri"`
}

func (e *Entry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias Entry
	var nodes struct {
		*Alias
		KElements []KElement `xml:"k_ele"`
		RElements []RElement `xml:"r_ele"`
	}

	nodes.Alias = (*Alias)(e)

	if err := d.DecodeElement(&nodes, &start); err != nil {
		return err
	}

	counter := 1
	for _, node := range nodes.KElements {
		if node.KePri == nil {
			node.KePri = make([]string, 0)
		}
		e.Readings = append(e.Readings, Reading{
			Word: node.Keb, Dictionary: node.KePri, IsKanji: true, CombinedId: e.EntryId*100 + counter,
		})
		counter++
	}

	for _, node := range nodes.RElements {
		if node.RePri == nil {
			node.RePri = make([]string, 0)
		}
		e.Readings = append(e.Readings, Reading{
			Word: node.Reb, Dictionary: node.RePri, IsKanji: false, CombinedId: e.EntryId*100 + counter,
		})
		counter++
	}

	return nil
}

func (e *Entry) String() string {
	var KReadings []string
	var WOKReadings []string
	for _, word := range e.Readings {
		if word.IsKanji {
			KReadings = append(KReadings, word.String())
		} else {
			WOKReadings = append(WOKReadings, word.String())
		}
	}
	wk := fmt.Sprintf(
		"Reading [incl. kanji]: %s",
		strings.Join(KReadings, ", "),
	)

	wok := fmt.Sprintf(
		"Reading [w/o. kanji]: %s",
		strings.Join(WOKReadings, ", "),
	)

	return fmt.Sprintf("%s\n%s", wk, wok)
}

func (e *Entry) IsPopular() bool {
	for _, r := range e.Readings {
		if r.IsInDictionary() {
			return true
		}
	}

	return false
}

func (e *Entry) IsInNews() bool {
	for _, word := range e.Readings {
		if word.NewsInDictionary() {
			return true
		}
	}
	return false
}

func (e *Entry) GetAllDictionaries() *[]string {
	dictMap := make(map[string]struct{})
	for _, word := range e.Readings {
		for _, dName := range word.Dictionary {
			dictMap[dName] = struct{}{}
		}
	}

	dictionaries := slices.Collect(maps.Keys(dictMap))

	return &dictionaries
}
