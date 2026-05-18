package parser

import (
	"encoding/xml"
	"errors"
	"io"
)

type Popular interface {
	IsPopular() bool
}

type XMLDictionary struct {
	Filepath string
	NodeName string
}

func CreateXMLParsingChan[T Popular](r io.Reader, nodeName string, cSize int) <-chan T {
	fh := xml.NewDecoder(r)
	fh.Strict = false

	kChan := make(chan T, cSize)

	go func() {
		var xmlSyntaxErr *xml.SyntaxError

		defer func() {
			close(kChan)
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
				if elem.Name.Local == nodeName {
					var node T
					err := fh.DecodeElement(&node, &elem)
					if err != nil {
						if errors.As(err, &xmlSyntaxErr) {
							// XML parser can't really parse modern name mapping
							// like &unc
						}
					}
					if node.IsPopular() {
						kChan <- node
					}
				}
			}
		}
	}()

	return kChan
}
