package parser

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/g13n4/LuteSentencePicker/tatoeba"
)

func CreateSudachiTSVParsingChan(tsvFileR io.Reader, cSize int) <-chan *tatoeba.SentenceTokens {
	reader := bufio.NewReader(tsvFileR)

	sChan := make(chan *tatoeba.SentenceTokens, cSize)
	re := regexp.MustCompile("(\\d+)\\s+jpn\\s+(.*)")

	go func() {
		defer func() {
			close(sChan)
		}()

		for {
			var sentence tatoeba.SentenceTokens
			line, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				panic(err)
			}
			if line == "" || line == "\n" {
				continue
			}
			line = strings.TrimSpace(line)
			for k, v := range re.FindStringSubmatch(line) {
				if k == 1 {
					sentence.Id, err = strconv.Atoi(v)
					if err != nil {
						panic(err)
					}
				}
				if k == 2 {
					split := strings.Split(v, " ")
					sentence.Tokens = &split
				}
			}
			sChan <- &sentence

		}
	}()

	return sChan
}
