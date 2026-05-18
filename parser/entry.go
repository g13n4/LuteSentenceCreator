//go:build ignore

package parser

import (
	"fmt"
	"log"
)

type EntryLiteral struct {
	EntryId int
}

type EntryDictionary struct {
	EntryId      int
	DictionaryId int
}

func CreateEntryToLiteralParsingChan(entryChan <-chan *string) <-chan *[]int {
	output := make(chan *[]int)
	for e := range entryChan {

		output <- &tokens
	}

	for s := range sentenceChan {
		_, err := fmt.Fprint(stdin, s)
		if err != nil {
			panic(err)
		}
	}

	err = stdin.Close()
	if err != nil {
		panic(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	return output
}
