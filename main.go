package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/g13n4/LuteSentencePicker/db"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/utils"
)

const DictionaryErrorMessage = "Value for %s is not set! You should provide a path to kanjidic, jmdict2 and japanese sentences"

func main() {
	stateSingleton := state.GetStateSingleton()

	dbStateRepo := repository.NewDBStateRepository(stateSingleton.Pool)
	currentDBStatus, err := dbStateRepo.GetStatus(context.Background())

	if err != nil {
		if strings.Contains(err.Error(), "relation \"db_state\" does not exist") {
			currentDBStatus = 0
		} else {
			panic(err)
		}
	}

	if currentDBStatus < 1 {
		sqlFile, err := os.ReadFile("./db/init.sql")
		if err != nil {
			panic(err)
		}

		fmt.Println("Initializing database...")
		_, err = stateSingleton.Pool.Exec(context.Background(), string(sqlFile))
		if err != nil {
			panic(err)
		}

		_, err = stateSingleton.Pool.Exec(context.Background(), "INSERT INTO db_state (id, status) VALUES (999, 0)")
		if err != nil {
			panic(err)
		}

		err = dbStateRepo.SetStatus(context.Background(), 1)

		if err != nil {
			panic(err)
		}
	}

	if currentDBStatus < 2 {
		kanjiPath := os.Getenv("PATH_KANJIDIC")
		if kanjiPath == "" {
			panic(fmt.Sprintf(DictionaryErrorMessage, "PATH_KANJIDIC"))
		}
		file, closer, err := utils.OpenFile(kanjiPath)
		defer func() {
			err := closer()
			if err != nil {
				panic(err)
			}
		}()
		if err != nil {
			panic(err)
		}

		fmt.Println("Loading kanji into DB...")
		err = db.FillKanji(stateSingleton, file)
		if err != nil {
			panic(err)
		}
		err = dbStateRepo.SetStatus(context.Background(), 2)

		if err != nil {
			panic(err)
		}
	}

	if currentDBStatus < 3 {
		jmdictPath := os.Getenv("PATH_JMDICT")
		if jmdictPath == "" {
			panic(fmt.Sprintf(DictionaryErrorMessage, "PATH_JMDICT"))
		}

		file, closer, err := utils.OpenFile(jmdictPath)
		defer func() {
			err := closer()
			if err != nil {
				panic(err)
			}
		}()
		if err != nil {
			panic(err)
		}

		fmt.Println("Loading words and readings into DB...")
		err = db.FillEntry(stateSingleton, file)
		if err != nil {
			panic(err)
		}
		err = dbStateRepo.SetStatus(context.Background(), 3)

		if err != nil {
			panic(err)
		}
	}

	if currentDBStatus < 4 {
		sentencePath := os.Getenv("PATH_SENTENCES")
		if sentencePath == "" {
			panic(fmt.Sprintf(DictionaryErrorMessage, "PATH_SENTENCES"))
		}

		senFile, closer, err := utils.OpenFile(sentencePath)
		defer func() {
			err := closer()
			if err != nil {
				panic(err)
			}
		}()
		if err != nil {
			panic(err)
		}

		sentenceSudachiPath := os.Getenv("PATH_SENTENCES_SUDACHI")
		if sentenceSudachiPath == "" {
			panic(fmt.Sprintf(DictionaryErrorMessage, "PATH_SENTENCES_SUDACHI"))
		}

		sudFile, closer, err := utils.OpenFile(sentenceSudachiPath)
		defer func() {
			err := closer()
			if err != nil {
				panic(err)
			}
		}()
		if err != nil {
			panic(err)
		}

		fmt.Println("Loading sentences into DB...")
		err = db.FillSentence(stateSingleton, senFile, sudFile)
		if err != nil {
			panic(err)
		}
		err = dbStateRepo.SetStatus(context.Background(), 4)

		if err != nil {
			panic(err)
		}
	}
}
