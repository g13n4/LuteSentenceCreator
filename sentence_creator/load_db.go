package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	db2 "github.com/g13n4/LuteSentencePicker/sentence_creator/db"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/parser"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

const DictionaryErrorMessage = "Value for %s is not set! You should provide a path to kanjidic, jmdict2 and japanese sentences"

func loadDB(stateSingleton *state.Singleton, status *atomic.Int64) {

	dbStateRepo := repository.NewDBStateRepository(stateSingleton.Pool)
	currentDBStatus, err := dbStateRepo.GetStatus(context.Background())

	sentencePath := os.Getenv("PATH_SENTENCES")
	if sentencePath == "" {
		panic(fmt.Sprintf(DictionaryErrorMessage, "PATH_SENTENCES"))
	}

	sentenceSudachiPath := os.Getenv("PATH_SENTENCES_SUDACHI")
	if sentenceSudachiPath == "" {
		panic(fmt.Sprintf(DictionaryErrorMessage, "PATH_SENTENCES_SUDACHI"))
	}

	sudachiDone := make(chan struct{})
	go func() {
		err := parser.ParseSentencesWithSudachi(sentencePath, sentenceSudachiPath)
		if err != nil {
			panic(err)
		}
		defer close(sudachiDone)
	}()

	if err != nil {
		if strings.Contains(err.Error(), "relation \"db_state\" does not exist") {
			currentDBStatus = 0
			status.Store(int64(currentDBStatus))
		} else {
			panic(err)
		}
	}

	if currentDBStatus < 2 {
		sqlFile, err := os.ReadFile("./db/init.sql")
		if err != nil {
			panic(err)
		}

		_, err = stateSingleton.Pool.Exec(context.Background(), string(sqlFile))
		if err != nil {
			panic(err)
		}

		_, err = stateSingleton.Pool.Exec(context.Background(), "INSERT INTO db_state (id, status) VALUES (999, 0)")
		if err != nil {
			panic(err)
		}

		currentDBStatus = 2
		err = dbStateRepo.SetStatus(context.Background(), currentDBStatus)
		status.Store(int64(currentDBStatus))

		if err != nil {
			panic(err)
		}
	}

	if currentDBStatus < 3 {
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

		err = db2.FillKanji(stateSingleton, file)
		if err != nil {
			panic(err)
		}

		currentDBStatus = 3
		err = dbStateRepo.SetStatus(context.Background(), currentDBStatus)
		status.Store(int64(currentDBStatus))

		if err != nil {
			panic(err)
		}
	}

	if currentDBStatus < 4 {
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

		err = db2.FillEntry(stateSingleton, file)
		if err != nil {
			panic(err)
		}

		currentDBStatus = 4
		err = dbStateRepo.SetStatus(context.Background(), currentDBStatus)
		status.Store(int64(currentDBStatus))

		if err != nil {
			panic(err)
		}
	}

	<-sudachiDone
	if currentDBStatus < 5 {
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

		err = db2.FillSentence(stateSingleton, senFile, sudFile)
		if err != nil {
			panic(err)
		}

		currentDBStatus = 5
		err = dbStateRepo.SetStatus(context.Background(), currentDBStatus)
		status.Store(int64(currentDBStatus))

		if err != nil {
			panic(err)
		}
	}

	status.Store(6)
}
