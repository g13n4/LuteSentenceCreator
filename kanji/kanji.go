package kanji

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

const kanjiNodeName = "character"

type Kanji struct {
	Literal string `xml:"literal"`

	JLPT      *int `xml:"misc>jlpt"`
	Frequency *int `xml:"misc>freq"`
	Grade     *int `xml:"misc>grade"`
}

func (k Kanji) String() string {
	var info []string

	if k.JLPT != nil {
		info = append(info, fmt.Sprintf("JLPT: %d", *k.JLPT))
	}

	if k.Frequency != nil {
		info = append(info, fmt.Sprintf("Frequency: %d", *k.Frequency))
	}

	if k.Grade != nil {
		info = append(info, fmt.Sprintf("Grade: %d", *k.Grade))
	}
	infoStr := strings.Join(info, ", ")

	if infoStr != "" {
		return fmt.Sprintf("%s: %s", k.Literal, infoStr)
	}

	return k.Literal
}

type Dictionary struct {
	Name  xml.Name `xml:"kanjidic2"`
	Kanji []Kanji  `xml:"character"`
}

func CreateKanjiChan(fn string, cSize int) <-chan Kanji {
	kanjiData, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	// offset xml file

	fh := xml.NewDecoder(kanjiData)
	kChan := make(chan Kanji, cSize)

	go func() {
		defer func() {
			close(kChan)
			err := kanjiData.Close()
			if err != nil {
				panic(err)
			}
		}()

		for {
			token, err := fh.Token()
			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}

			if token == nil {
				continue
			}

			switch elem := token.(type) {
			case xml.StartElement:
				if elem.Name.Local == kanjiNodeName {
					var kNode Kanji
					if err := fh.DecodeElement(&kNode, &elem); err != nil {
						fmt.Println(err)
					}
					kChan <- kNode
				}

			}
		}
	}()

	return kChan
}
