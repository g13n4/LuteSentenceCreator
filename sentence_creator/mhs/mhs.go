package mhs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
	"github.com/jackc/pgx/v5"
)

type Executor struct {
	db repository.DBSaver
}

func NewExecutor(db repository.DBSaver) *Executor {
	return &Executor{db: db}
}

func (exc *Executor) GetSentences(ctx context.Context, outputFile *os.File, mshq QueryHelper, permuts, limit int) error {
	var sMap *map[int64]*[]int64
	var err error

	log.Println("Query sentences")

	sMap, err = exc.getSentencesIdRows(ctx, mshq.CreateQuery(), mshq.GetPreallocSize())

	if err != nil {
		return err
	}
	log.Println("Filling input with data")
	fileNameIn, fileNameOut := exc.getTemporaryFilePath()
	err = exc.createMHSFileInput(fileNameIn, sMap)
	if err != nil {
		return err
	}
	log.Println("Run graph optimization")
	sentenceIds, err := exc.processSentenceSetFile(fileNameIn, fileNameOut, permuts)
	if err != nil {
		return err
	}

	log.Println("Creating output")
	return exc.saveSentencesSet(ctx, outputFile, sentenceIds, limit)
}

func (exc *Executor) getSentencesIdRows(ctx context.Context, sqlQuery string, prealocSize int) (*map[int64]*[]int64, error) {
	rows, err := exc.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return exc.createSentenceSet(rows, prealocSize)
}

func (exc *Executor) createSentenceSet(rows pgx.Rows, prealocSize int) (*map[int64]*[]int64, error) {
	rToS := make(map[int64]*[]int64, prealocSize)

	var readingId int64
	var sentenceId int64
	for rows.Next() {
		err := rows.Scan(&readingId, &sentenceId)
		if err != nil {
			return nil, err
		}

		_, ok := rToS[readingId]
		if ok {
			*rToS[readingId] = append(*rToS[readingId], sentenceId)
		} else {
			rToS[readingId] = &[]int64{sentenceId}
		}
	}

	return &rToS, nil
}

func (exc *Executor) getTemporaryFilePath() (string, string) {
	tempPrefix := fmt.Sprintf("%v", time.Now().Unix())
	tempNameInput := fmt.Sprintf("input-%v.dat", tempPrefix)
	tempNameOutput := fmt.Sprintf("output-%v.dat", tempPrefix)

	mhsInput := filepath.Join(os.TempDir(), tempNameInput)
	mhsOutputName := filepath.Join(os.TempDir(), tempNameOutput)
	return mhsInput, mhsOutputName
}

func (exc *Executor) createMHSFileInput(mhsInput string, sentenceMap *map[int64]*[]int64) error {
	file, err := os.Create(mhsInput)
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return err
	}

	writer := bufio.NewWriterSize(file, 64*1024)

	buffrdLine := make([]byte, 0, 1024)

	for _, line := range *sentenceMap {
		buffrdLine = buffrdLine[:0]
		for _, num := range *line {
			buffrdLine = strconv.AppendInt(buffrdLine, num, 10)
			buffrdLine = append(buffrdLine, ' ')
		}
		buffrdLine = append(buffrdLine, '\n')
		if _, err := writer.Write(buffrdLine); err != nil {
			return err
		}
	}

	return nil
}

func (exc *Executor) processSentenceSetFile(input, output string, permuts int) (<-chan *string, error) {
	mhsExec := os.Getenv("MHS_PATH")
	if mhsExec == "" {
		return nil, errors.New("MHS_PATH environment variable is not set")
	}

	cmd := exec.Command(
		"/mhs/agdmhs",
		fmt.Sprintf("--input=%s", input),
		fmt.Sprintf("--output=%s", output),
		"-a",
		"pmmcs",
	)
	outputChan := make(chan *string)

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Command failed: %s\nOutput: %s", err, cmdOutput))
	}

	file, closer, err := utils.OpenFile(output)

	go func() {
		defer func() {
			err := closer()
			if err != nil {
				panic(err)
			}
		}()

		fScanner := bufio.NewScanner(file)
		sentenceSet := make(map[string]struct{})

		lineCounter := 0
		for fScanner.Scan() {
			line := fScanner.Text()
			line = strings.TrimSpace(line)
			numbers := strings.Split(line, " ")
			for _, num := range numbers {
				log.Println("val", num)
				_, ok := sentenceSet[num]
				if ok {
					continue
				}

				outputChan <- &num
				sentenceSet[num] = struct{}{}
			}
			lineCounter++

			if lineCounter >= permuts {
				close(outputChan)
				return
			}
		}

		close(outputChan)
		return
	}()

	return outputChan, err
}

func (exc *Executor) saveSentencesSet(ctx context.Context, outputFile *os.File, sentenceIdChan <-chan *string, limit int) error {
	ids := make([]string, 0, 100)

	for sentenceId := range sentenceIdChan {
		ids = append(ids, *sentenceId)
	}

	query := fmt.Sprintf("SELECT sentence from sentences WHERE id = ANY($1) LIMIT %v", limit)
	rows, err := exc.db.Query(ctx, query, ids)

	if err != nil {
		return err
	}

	var sentence string
	writer := bufio.NewWriter(outputFile)
	for rows.Next() {
		err := rows.Scan(&sentence)
		if err != nil {
			return err
		}
		_, err = writer.WriteString(sentence + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	return err
}

func (exc *Executor) collectSentenceSet(ctx context.Context, sentenceIdChan <-chan *string, limit int) ([]string, error) {
	ids := make([]string, 0)

	for sentenceId := range sentenceIdChan {
		ids = append(ids, *sentenceId)
	}

	query := fmt.Sprintf("SELECT sentence from sentences WHERE id = ANY($1) LIMIT %v", limit)
	rows, err := exc.db.Query(ctx, query, ids)

	sentences := make([]string, 0)
	if err != nil {
		return sentences, err
	}

	for rows.Next() {
		var sentence string

		err := rows.Scan(&sentence)
		if err != nil {
			return nil, err
		}

		sentences = append(sentences, sentence)
	}

	return sentences, nil
}
