package main

import (
	"fmt"
	"path/filepath"
	"time"

	kanji "github.com/g13n4/LuteSentenceCreator/kanji"
)

func main() {
	fName, err := filepath.Abs("./resources/kanjidic2-test.xml")
	if err != nil {
		panic(err)
	}
	kChan := kanji.CreateKanjiChan(fName, 10)
	time.Sleep(1 * time.Second)
	var c int = 1

	for k := range kChan {
		fmt.Println(c, k)
		c += 1
		if c == 10 {
			break
		}
	}
}
