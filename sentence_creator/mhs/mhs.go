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
	"sync"
	"time"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

type Executor struct {
	db                 repository.DBSaver
	maxInputLinesCount int
	maxEdgesCount      int
	maxPermutations    int
}

type SentenceReadingRow struct {
	SentenceId int64
	ReadingId  int64
}

func NewExecutor(db repository.DBSaver) *Executor {
	var inputLines int = 64
	maxSize := os.Getenv("MHS_MAX_SLICE_SIZE")
	if maxSize == "" {
		log.Print("MHS_MAX_SLICE_SIZE environment variable is not set. Using default")
	} else {
		val, err := strconv.Atoi(maxSize)
		if err != nil {
			log.Print("MHS_MAX_SLICE_SIZE environment variable is not correct. Using default")
		} else {
			inputLines = val
		}
	}

	return &Executor{db: db, maxInputLinesCount: inputLines, maxEdgesCount: inputLines * 2, maxPermutations: 1}
}

func (exc *Executor) GetSentences(ctx context.Context, outputFile *os.File, mshq QueryHelper, limit int) error {
	var wg sync.WaitGroup

	outputSentenceIdChan := make(chan *string)

	err := exc.processSentencesIdRows(ctx, mshq.CreateQuery(), mshq.GetPreallocSize(), outputSentenceIdChan, &wg)
	if err != nil {
		return err
	}

	go func() {
		wg.Wait()
		close(outputSentenceIdChan)
	}()

	return exc.saveSentencesSet(ctx, outputFile, outputSentenceIdChan, limit)
}

func (exc *Executor) getIDs(ctx context.Context, sqlQuery string) (chan SentenceReadingRow, error) {
	rows, err := exc.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	output := make(chan SentenceReadingRow)

	go func() {
		defer rows.Close()

		var readingId int64
		var sentenceId int64

		for rows.Next() {
			err := rows.Scan(&readingId, &sentenceId)
			if err != nil {
				panic(err)
			}

			output <- SentenceReadingRow{SentenceId: sentenceId, ReadingId: readingId}
		}
	}()

	return output, nil
}

func (exc *Executor) processSentencesIdRows(
	ctx context.Context,
	sqlQuery string,
	preallocSize int,
	outputSentenceIdChan chan<- *string,
	wg *sync.WaitGroup,
) error {
	rowsChan, err := exc.getIDs(ctx, sqlQuery)
	if err != nil {
		return err
	}

	rToS := make(map[int64]*[]int64, preallocSize)

	var mu sync.Mutex

	var prevReadingId int64
	var curRowCounter int
	var totalRowCounter int

	var canBeSliced bool

	// We use first element to ensure there are at least two different sets in MHSA
	var firstElementId int64
	var firstElementList = make([]int64, 0)

	for srr := range rowsChan {
		curRowCounter++
		canBeSliced = prevReadingId != srr.ReadingId

		if srr.ReadingId == firstElementId {
			firstElementList = append(firstElementList, srr.ReadingId)
		}

		_, ok := rToS[srr.ReadingId]
		if ok {
			*rToS[srr.ReadingId] = append(*rToS[srr.ReadingId], srr.SentenceId)
		} else {
			rToS[srr.ReadingId] = &[]int64{srr.SentenceId}
		}

		if canBeSliced && curRowCounter >= exc.maxInputLinesCount {
			totalRowCounter += curRowCounter
			log.Printf("Sent %v to process in MHS. Total rows send - %v", curRowCounter, totalRowCounter)
			wg.Go(func() {
				err := exc.processMapInput(rToS, outputSentenceIdChan, &mu)
				if err != nil {
					panic(err)
				}
			})

			rToS = make(map[int64]*[]int64, preallocSize)
			curRowCounter = 0
		}

		prevReadingId = srr.ReadingId
	}

	if len(rToS) != 0 {
		if len(rToS) == 1 {
			_, ok := rToS[firstElementId]
			if !ok {
				rToS[firstElementId] = &firstElementList
			}
		}

		totalRowCounter += curRowCounter
		log.Printf("Sent %v to process in MHS. Total rows send - %v", curRowCounter, totalRowCounter)
		wg.Go(func() {
			err := exc.processMapInput(rToS, outputSentenceIdChan, &mu)
			if err != nil {
				panic(err)
			}
		})
	}
	return nil
}

func (exc *Executor) processMapInput(
	sentenceMap map[int64]*[]int64,
	outputSentenceIdChan chan<- *string,
	mu *sync.Mutex,
) error {
	fileNameIn := exc.getTemporaryFilePath("input")
	err := exc.createMHSFileInput(fileNameIn, sentenceMap)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	fileNameOut := exc.getTemporaryFilePath("output")
	err = exc.processSentenceSetFile(fileNameIn, fileNameOut, outputSentenceIdChan)
	return err
}

func (exc *Executor) getTemporaryFilePath(prefix string) string {
	curTime := fmt.Sprintf("%v", time.Now().UnixNano())
	tempNameInput := fmt.Sprintf("%s-temp-%v.dat", prefix, curTime)

	tempFilePath := filepath.Join(os.TempDir(), tempNameInput)
	return tempFilePath
}

func (exc *Executor) createMHSFileInput(mhsInput string, sentenceMap map[int64]*[]int64) error {
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

	writer := bufio.NewWriterSize(file, 2*1024)

	bufferedLine := make([]byte, 0, 2*1024)

	for _, line := range sentenceMap {
		bufferedLine = bufferedLine[:0]
		for _, num := range *line {
			bufferedLine = strconv.AppendInt(bufferedLine, num, 10)
			bufferedLine = append(bufferedLine, ' ')
		}
		bufferedLine = append(bufferedLine, '\n')

		_, err := writer.Write(bufferedLine)
		if err != nil {
			return err
		}

		err = writer.Flush()
		if err != nil {
			return err
		}

	}

	return nil
}

func (exc *Executor) processSentenceSetFile(
	inFileName string,
	outFileName string,
	outputSentenceIdChan chan<- *string,
) error {
	mhsThreadsCount := os.Getenv("MHS_THREADS")
	if mhsThreadsCount == "" {
		mhsThreadsCount = "1"
	}

	log.Printf("Started MHS processing")
	cmd := exec.Command(
		"/mhs/agdmhs",
		fmt.Sprintf("--input=%s", inFileName),
		fmt.Sprintf("--output=%s", outFileName),
		"-t",
		mhsThreadsCount,
		"-a",
		"pmmcs",
	)

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(
			fmt.Sprintf("Command failed:\nfile input:%s\nfile output:%s\n%s\nOutput: %s",
				inFileName,
				outFileName,
				err,
				cmdOutput,
			))
	}

	file, closer, err := utils.OpenFile(outFileName)
	if err != nil {
		return errors.New(fmt.Sprintf("Error opening output file: %s", err))
	}

	defer func() {
		err := closer()
		if err != nil {
			panic(err)
		}

		// TODO: uncomment
		//err = os.Remove(inFileName)
		//if err != nil {
		//	panic(err)
		//}
		//
		//err = os.Remove(outFileName)
		//if err != nil {
		//	panic(err)
		//}
	}()

	fScanner := bufio.NewScanner(file)

	lineCount := 0
	sentenceCount := 0
	for fScanner.Scan() {
		line := fScanner.Text()
		line = strings.TrimSpace(line)
		numbers := strings.Split(line, " ")
		for _, num := range numbers {
			sentenceCount++
			outputSentenceIdChan <- &num
		}
		lineCount++

		if lineCount >= exc.maxPermutations {
			break
		}
	}

	log.Printf("Extracted %v sentences from MHS", sentenceCount)

	return nil
}

func (exc *Executor) saveSentencesSet(ctx context.Context, outputFile *os.File, sentenceIdChan <-chan *string, limit int) error {
	ids := make([]string, 0, 1000)

	sentenceSet := make(map[string]struct{})
	for sentenceId := range sentenceIdChan {
		_, ok := sentenceSet[*sentenceId]
		if ok {
			continue
		}

		ids = append(ids, *sentenceId)
		sentenceSet[*sentenceId] = struct{}{}
	}

	var sqlLimit string
	if limit != 0 {
		sqlLimit = fmt.Sprintf("LIMIT %v", limit)
	}

	query := fmt.Sprintf("SELECT sentence from sentences WHERE id = ANY($1) %s", sqlLimit)
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
