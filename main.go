package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/g13n4/LuteSentencePicker/db"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/utils"
	"github.com/joho/godotenv"
)

const DictionaryErrorMessage = "Value for %s is not set! You should provide a path to kanjidic, jmdict2 and japanese sentences"

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	stateSingleton := state.GetStateSingleton()

	dbStateRepo := repository.NewDBStateRepository(stateSingleton.Pool)
	currentDBStatus, err := dbStateRepo.GetStatus(context.Background())

	if err != nil {
		panic(err)
	}

	if currentDBStatus < 1 {
		sqlFile, err := os.ReadFile("init.sql")

		if err != nil {
			panic(err)
		}
		_, err = stateSingleton.Pool.Exec(context.Background(), string(sqlFile))

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
		if kanjiPath != "" {
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
		if jmdictPath != "" {
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
		if sentencePath != "" {
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
		if sentenceSudachiPath != "" {
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
