package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/g13n4/LuteSentenceCreator/jmdict"
	"github.com/g13n4/LuteSentenceCreator/kanji"

	"github.com/g13n4/LuteSentenceCreator/parser"
)

func main() {
	kanjiDictFileName, err := filepath.Abs("./resources/kanjidic2.xml")
	if err != nil {
		panic(err)
	}
	kanjiDictObj := parser.XMLDictionary{
		Filepath: kanjiDictFileName,
		NodeName: kanji.KanjiNodeName,
	}

	kChan := parser.CreateParsingChan[kanji.Kanji](kanjiDictObj, 10)

	time.Sleep(1 * time.Second)
	var c int = 1

	for k := range kChan {
		fmt.Println(c, k)
		c += 1
		if c == 10 {
			break
		}
	}

	entryDictFileName, err := filepath.Abs("./resources/JMDict.xml")
	if err != nil {
		panic(err)
	}
	entryDictObj := parser.XMLDictionary{
		Filepath: entryDictFileName,
		NodeName: jmdict.EntryNodeName,
	}

	eChan := parser.CreateParsingChan[jmdict.Entry](entryDictObj, 10)
	ec := 1

	for k := range eChan {
		fmt.Println(ec, k)
		ec += 1
		if ec == 10 {
			break
		}
	}
}
