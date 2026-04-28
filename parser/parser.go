package parser

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/g13n4/LuteSentencePicker/tatoeba"
)

type Popular interface {
	IsPopular() bool
}

type XMLDictionary struct {
	Filepath string
	NodeName string
}

func CreateXMLParsingChan[T Popular](xmlData XMLDictionary, cSize int) <-chan T {
	file, err := os.Open(xmlData.Filepath)
	if err != nil {
		panic(err)
	}
	// offset xml file

	fh := xml.NewDecoder(file)
	fh.Strict = false

	kChan := make(chan T, cSize)

	go func() {
		var xmlSyntaxErr *xml.SyntaxError

		defer func() {
			close(kChan)
			err := file.Close()
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

func CreateTSVParsingChan(tsvFilePath string, cSize int) <-chan tatoeba.Sentence {
	file, err := os.Open(tsvFilePath)
	if err != nil {
		panic(err)
	}
	// offset xml file

	reader := bufio.NewReader(file)

	sChan := make(chan tatoeba.Sentence, cSize)

	go func() {
		defer func() {
			close(sChan)
			err := file.Close()
			if err != nil {
				panic(err)
			}
		}()

		for {
			var sentence tatoeba.Sentence
			line, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				panic(err)
			}
			line = strings.TrimSpace(line)
			split := strings.Split(line, "\t")
			sentenceId, err := strconv.Atoi(split[0])
			if err != nil {
				panic(fmt.Errorf("wrong tsv file format. parser expects [sentence_id]\\tjpn\\s[sentence]: %w", err))
			}
			sentence.Id = sentenceId
			sentence.Text = split[2]
			sChan <- sentence

		}
	}()

	return sChan
}
