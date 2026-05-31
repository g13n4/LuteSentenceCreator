package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
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

func ParseSentencesWithSudachi(inputFile, outputFile string) error {
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return errors.New("sentence file does not exist. You can download one from tatoeba")
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		sentenceSudachiPath := os.Getenv("SUDACHI_PATH")
		if sentenceSudachiPath == "" {
			return errors.New("can not find sudachi executable")
		}

		err := exec.Command(
			fmt.Sprintf("%s -w --split-sentences no %s > %s",
				sentenceSudachiPath,
				inputFile,
				outputFile)).Run()

		if err != nil {
			return err
		}

		return nil
	}
	return nil
}
