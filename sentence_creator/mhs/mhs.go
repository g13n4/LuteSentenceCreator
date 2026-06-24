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

	ignoreErrors   bool
	errorThreshold float32
}

type SentenceReadingRow struct {
	SentenceId int64
	ReadingId  int64
}

func NewExecutor(db repository.DBSaver, maxSlice ...int) *Executor {
	var inputLines int
	if len(maxSlice) != 0 {
		inputLines = maxSlice[0]
	} else {
		inputLines = utils.GetEnvIntValue("MHS_MAX_SLICE_SIZE", 700)
	}

	return &Executor{
		db:                 db,
		maxInputLinesCount: inputLines,
		maxEdgesCount:      inputLines * 2,
		maxPermutations:    1,
		ignoreErrors:       true,
		errorThreshold:     0.,
	}
}

func (exc *Executor) GetSentences(ctx context.Context, outputFile *os.File, mshq QueryHelper, limit int) error {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	outputSentenceIdChan := make(chan *string)
	processingErrors := make(chan error, 128)

	rowsChan, err := exc.getIDs(ctx, mshq.CreateQuery())
	if err != nil {
		return err
	}

	err = exc.processSentencesIdRows(ctx, &wg, mshq.GetPreallocSize(), rowsChan, outputSentenceIdChan, processingErrors)
	if err != nil {
		return err
	}

	go func() {
		wg.Wait()
		close(outputSentenceIdChan)
	}()

	var errCounter, elementCounter int
	sentenceSet := make(map[string]struct{})

outerLoop:
	for {
		select {
		case err := <-processingErrors:
			if err != nil {
				if !exc.ignoreErrors {
					cancel()
					return err
				}
				errCounter++
			}
			elementCounter++
		case sentenceId, ok := <-outputSentenceIdChan:
			if !ok {
				break outerLoop
			}

			_, ok = sentenceSet[*sentenceId]
			if ok {
				continue
			}

			sentenceSet[*sentenceId] = struct{}{}
		}
	}

	if elementCounter != 0 && errCounter != 0 && exc.errorThreshold != 0. {
		errPercent := float32(errCounter) / float32(elementCounter)
		if errPercent > exc.errorThreshold {
			return errors.New(fmt.Sprintf("too many errors in during processing: expected %v and got %v (%v errors out of %v goroutines)", exc.errorThreshold, errPercent, errCounter, elementCounter))
		}
	}

	return exc.saveSentencesSet(ctx, outputFile, &sentenceSet, limit)
}

func (exc *Executor) getIDs(ctx context.Context, sqlQuery string) (chan SentenceReadingRow, error) {
	rows, err := exc.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	output := make(chan SentenceReadingRow, 100)

	go func() {
		defer close(output)
		defer rows.Close()

		var readingId int64
		var sentenceId int64
		var found bool

		for rows.Next() {
			found = true

			err := rows.Scan(&readingId, &sentenceId)
			if err != nil {
				panic(err)
			}
			output <- SentenceReadingRow{SentenceId: sentenceId, ReadingId: readingId}
		}
		if !found {
			panic("no rows found")
		}
	}()

	return output, nil
}

func (exc *Executor) processSentencesIdRows(
	ctx context.Context,
	wg *sync.WaitGroup,
	preallocSize int,
	queryChan <-chan SentenceReadingRow,
	outputSentenceIdChan chan<- *string,
	errorsChan chan<- error,
) error {

	rToS := make(map[int64]*[]int64, preallocSize)

	var mu sync.Mutex

	var workerId int

	var prevReadingId int64
	var elementCounter int
	var totalElementCounter int

	var canBeSliced bool

	// We use first element to ensure there are at least two different sets in MHSA
	var firstElementId int64
	var firstElementList = make([]int64, 0)

	parallelepiped := func(sentenceMap map[int64]*[]int64, w int, lastProcess bool) {
		defer wg.Done()
		if lastProcess {
			log.Printf("[worker %v] Sent last %v lines to MHS. Total elements send - %v", w, len(sentenceMap), totalElementCounter)
		} else {
			log.Printf("[worker %v] Sent %v lines to MHS. Total elements send - %v", w, len(sentenceMap), totalElementCounter)
		}
		err := exc.processMapInput(ctx, &sentenceMap, outputSentenceIdChan, &mu, w)

		if err != nil {
			errorsChan <- errors.New(fmt.Sprintf("[worker %v] error: \n%+v", w, err))
		}
	}

	for srr := range queryChan {
		elementCounter++
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

		if canBeSliced && len(rToS) >= exc.maxInputLinesCount {
			totalElementCounter += elementCounter
			workerId++

			wg.Add(1)
			go parallelepiped(rToS, workerId, false)

			rToS = make(map[int64]*[]int64, preallocSize)
			elementCounter = 0
		}

		prevReadingId = srr.ReadingId
	}

	if len(rToS) != 0 {
		_, ok := rToS[firstElementId]
		if !ok {
			rToS[firstElementId] = &firstElementList
		}

		totalElementCounter += elementCounter

		wg.Add(1)
		go parallelepiped(rToS, workerId+1, true)
	}

	return nil
}

func (exc *Executor) processMapInput(
	ctx context.Context,
	sentenceMap *map[int64]*[]int64,
	outputSentenceIdChan chan<- *string,
	mu *sync.Mutex,
	t int,
) error {
	fileNameIn := exc.getTemporaryFilePath("input")
	err := exc.createMHSFileInput(fileNameIn, sentenceMap)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(fileNameIn)
	}()

	mu.Lock()
	defer mu.Unlock()

	select {
	case <-ctx.Done():
		return nil
	default:
		fileNameOut := exc.getTemporaryFilePath("output")
		defer func() {
			_ = os.Remove(fileNameOut)
		}()

		err = exc.processSentenceSetFile(fileNameIn, fileNameOut, outputSentenceIdChan, t)
		return err
	}
}

func (exc *Executor) getTemporaryFilePath(prefix string) string {
	curTime := fmt.Sprintf("%v", time.Now().UnixNano())
	tempNameInput := fmt.Sprintf("%s-temp-%v.dat", prefix, curTime)

	tempFilePath := filepath.Join(os.TempDir(), tempNameInput)
	return tempFilePath
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

	bufferSize := 1024

	writer := bufio.NewWriterSize(file, bufferSize)

	bufferedLine := make([]byte, 0, bufferSize)

	for _, line := range *sentenceMap {
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
	t int,
) error {
	mhsThreadsCount := os.Getenv("MHS_THREADS")
	if mhsThreadsCount == "" {
		mhsThreadsCount = "1"
	}

	log.Printf("[worker %v] Started MHS processing", t)
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
			fmt.Sprintf("[worker %v] Command failed:\nfile input:%s\nfile output:%s\n%s\nOutput: %s",
				t,
				inFileName,
				outFileName,
				err,
				cmdOutput,
			))
	}
	fmt.Printf("%s", cmdOutput)

	file, closer, err := utils.OpenFile(outFileName)
	if err != nil {
		return errors.New(fmt.Sprintf("[worker %v] Error opening output file: %s", t, err))
	}

	defer func() {
		err := closer()
		if err != nil {
			panic(err)
		}

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

	log.Printf("[worker %v] Extracted %v sentences from MHS", t, sentenceCount)

	return nil
}

func (exc *Executor) saveSentencesSet(
	ctx context.Context,
	outputFile *os.File,
	sentenceSet *map[string]struct{},
	limit int,
) error {
	ids := make([]string, 0, 1000)

	for sentenceId := range *sentenceSet {
		ids = append(ids, sentenceId)
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

func (exc *Executor) SetBehaviorOnError(ignoreErrors bool) {
	exc.ignoreErrors = ignoreErrors
}

func (exc *Executor) SetErrorThreshold(errorThreshold float32) {
	exc.errorThreshold = errorThreshold
}
