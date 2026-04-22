package parser

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
)

type Popular interface {
	IsPopular() bool
}

type XMLDictionary struct {
	Filepath string
	NodeName string
}

func CreateParsingChan[T Popular](xmlData XMLDictionary, cSize int) <-chan T {
	kanjiData, err := os.Open(xmlData.Filepath)
	if err != nil {
		panic(err)
	}
	// offset xml file

	fh := xml.NewDecoder(kanjiData)
	fh.Strict = false

	kChan := make(chan T, cSize)

	go func() {
		var xmlSyntaxErr *xml.SyntaxError

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
				if errors.As(err, &xmlSyntaxErr) {
					continue
				}
				panic(err)
			}

			if token == nil {
				continue
			}

			switch elem := token.(type) {
			case xml.StartElement:
				if elem.Name.Local == xmlData.NodeName {
					var node T
					err := fh.DecodeElement(&node, &elem)
					if err != nil {
						if errors.As(err, &xmlSyntaxErr) {
							// XML parser can't really parse modern name mapping
							// like &unc
						}
					}
					// We only need "popular" words and kanji
					//fmt.Println(node)
					if node.IsPopular() {
						kChan <- node
					}
				}
			}
		}
	}()

	return kChan
}
