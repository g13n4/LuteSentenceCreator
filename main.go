package main

import (
	//	"context"
	"fmt"
	//	"os"
	"path/filepath"
	"time"

	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/kanji"
	//	"github.com/jackc/pgx/v5"

	"github.com/g13n4/LuteSentencePicker/parser"
)

func main() {
	//Url := fmt.Sprintf(
	//	"postgres://%s:%s@%s:%s/%s?sslmode=disable",
	//	os.Getenv("DATABASE_USERNAME"),
	//	os.Getenv("DATABASE_PASSWORD"),
	//	os.Getenv("DATABASE_ADDRESS"),
	//	os.Getenv("DATABASE_PORT"),
	//	os.Getenv("DATABASE_NAME"),
	//)
	//conn, err := pgx.Connect(context.Background(), Url)
	//defer func() {
	//	err := conn.Close(context.Background())
	//	panic(err)
	//}()
	//
	//if err != nil {
	//	panic(err)
	//}

	kanjiDictFileName, err := filepath.Abs("./resources/kanjidic2.xml")
	if err != nil {
		panic(err)
	}
	kanjiDictObj := parser.XMLDictionary{
		Filepath: kanjiDictFileName,
		NodeName: kanji.KanjiNodeName,
	}

	kChan := parser.CreateXMLParsingChan[*kanji.Kanji](kanjiDictObj, 10)

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

	eChan := parser.CreateXMLParsingChan[*jmdict.Entry](entryDictObj, 10)
	ec := 1

	for k := range eChan {
		fmt.Println(ec, k)
		ec += 1
		if ec == 10 {
			break
		}
	}
}
