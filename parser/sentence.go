package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/g13n4/LuteSentencePicker/tatoeba"
)

func CreateTSVParsingChan(tsvR io.Reader, cSize int) <-chan *tatoeba.Sentence {
	reader := bufio.NewReader(tsvR)

	sChan := make(chan *tatoeba.Sentence, cSize)

	go func() {
		defer func() {
			close(sChan)
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
			sChan <- &sentence

		}
	}()

	return sChan
}
