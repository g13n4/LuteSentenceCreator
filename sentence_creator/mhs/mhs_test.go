package mhs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
)

const testWord = "あたたかい"

func TestCreateMHSFileInput(t *testing.T) {
	stateSingleton := state.GetStateSingleton()

	mhsExecutor := NewExecutor(stateSingleton.Pool)

	mhsInput := filepath.Join(os.TempDir(), "test-intput.dat")
	defer func() {
		err := os.Remove(mhsInput)
		if err != nil {
			t.Errorf("error removing input file: %+v", err)
		}
	}()

	testMapPreallocSize := 10_000
	testMapTotalElementSize := 1_000_000
	testMapValueSize := 1_000_000 / testMapPreallocSize

	testMap := make(map[int64]*[]int64, testMapPreallocSize)

	var keyValue int64 = 1
	for i := 1; i < testMapTotalElementSize; i++ {
		if i%testMapValueSize == 0 {
			keyValue++
		}
		i64 := int64(i)
		_, ok := testMap[keyValue]
		if !ok {
			testMap[keyValue] = &[]int64{i64}
		} else {
			*testMap[keyValue] = append(*testMap[keyValue], i64)
		}
	}

	err := mhsExecutor.createMHSFileInput(mhsInput, &testMap)
	if err != nil {
		t.Errorf("Error scanning row: %v", err)
	}

	f, err := os.Stat(mhsInput)
	if os.IsNotExist(err) {
		t.Errorf("didn't create file")
	}

	// both numbers and spaces are in ascii character map so their size is only 1 byte
	expectedSize := int64(testMapTotalElementSize + testMapTotalElementSize)
	if f.Size() < expectedSize {
		t.Errorf("file size is smaller than expected. file size: %v, expected size: %v", f.Size(), expectedSize)
	}

}

// go test -run ^TestProcessSentenceSetFile ./mhs
func TestProcessSentenceSetFile(t *testing.T) {
	var wg sync.WaitGroup
	stateSingleton := state.GetStateSingleton()

	mhsExecutor := NewExecutor(stateSingleton.Pool)

	mhsInput := filepath.Join(os.TempDir(), "test-intput.dat")
	mhsOutput := filepath.Join(os.TempDir(), "test-output.dat")

	testExample := []byte(`1 2 5 
2 3 4 
1 3 `)

	err := os.WriteFile(mhsInput, testExample, 0777)

	if err != nil {
		t.Errorf("can't write to a file in a directory %v", err)
	}

	testChan := make(chan *string)

	wg.Go(func() {
		err = mhsExecutor.processSentenceSetFile(mhsInput, mhsOutput, testChan, 1)
		if err != nil {
			t.Errorf("error using mhsa exec %v", err)
		}
	})

	go func() {
		wg.Wait()
		close(testChan)
	}()

	count := 0
	var output string

	for v := range testChan {
		count++
		if output != "" {
			output += " " + *v
		} else {
			output += *v
		}
	}

	if 1 >= count || count > 3 {
		t.Errorf("expected at least 2 numbers in a set and got %v\noutput: %s", count, output)
	}
	if output == "" {
		t.Errorf("output is empty")
	}
}

// go test -v -run ^TestSentenceIdRowsProcessing ./mhs
func TestSentenceIdRowsProcessing(t *testing.T) {
	var wg sync.WaitGroup
	stateSingleton := state.GetStateSingleton()

	mhsExecutor := NewExecutor(stateSingleton.Pool, 3)

	testExample := [][]int{
		{1, 2, 5, 4},
		{2, 3, 4, 6, 7},
		{1, 3, 7, 2},
		{5, 2, 1, 8},
		{8, 9, 5, 3, 2, 1},
		{2, 5, 6, 7, 8},
	}

	vChan := make(chan SentenceReadingRow, 100)
	go func() {
		defer close(vChan)

		for idx, vSlice := range testExample {
			for _, v := range vSlice {
				vChan <- SentenceReadingRow{SentenceId: int64(v), ReadingId: int64(idx)}
			}
		}
	}()

	testChan := make(chan *string)

	errChan := make(chan error, 10)
	err := mhsExecutor.processSentencesIdRows(context.Background(), &wg, 10, vChan, testChan, errChan)
	if err != nil {
		t.Errorf("error using mhsa exec %v", err)
	}

	close(errChan)
	for err := range errChan {
		if err != nil {
			t.Errorf("error in a processing goroutine %v", err)
		}
	}

	go func() {
		wg.Wait()
		close(testChan)
	}()

	count := 0
	var output string

	for v := range testChan {
		count++
		if output != "" {
			output += " " + *v
		} else {
			output += *v
		}
	}

	if count < 3 {
		t.Errorf("expected at least 2 numbers in a set and got %v\noutput: %s", count, output)
	}
}

func TestCreateSentenceSet(t *testing.T) {
	stateSingleton := state.GetStateSingleton()

	mhsExecutor := NewExecutor(stateSingleton.Pool)

	limitToSubquery := fmt.Sprintf("WHERE  smr.s_id in (SELECT DISTINCT smr.s_id from sentences__mtm__readings smr JOIN readings r ON smr.r_id = r.id JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry where r.reading = '%s')", testWord)

	// test without DISTINCT to make query less complicated
	querySQL := "SELECT smr.r_id, smr.s_id from sentences__mtm__readings smr JOIN readings r ON smr.r_id = r.id JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry " + limitToSubquery
	groupedSQL := "SELECT smr.r_id, count(smr.s_id) from sentences__mtm__readings smr JOIN readings r ON smr.r_id = r.id JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry " + limitToSubquery + " GROUP BY smr.r_id"

	rows, err := mhsExecutor.getIDs(context.Background(), querySQL)
	if err != nil {
		t.Errorf("Error executing query: %s\n%v", querySQL, err)
	}

	outMapSize := make(map[int64]int64)
	for row := range rows {
		outMapSize[row.ReadingId]++
	}

	groupedRows, err := stateSingleton.Pool.Query(context.Background(), groupedSQL)
	if err != nil {
		t.Errorf("Error executing grouped query: : %s\n%v", groupedSQL, err)
	}

	var rowId, groupedRowSize int64
	for groupedRows.Next() {
		err := groupedRows.Scan(&rowId, &groupedRowSize)
		if err != nil {
			t.Errorf("Error scanning row: %v", err)
		}

		rowSize := outMapSize[rowId]
		if outMapSize[rowId] != groupedRowSize {
			t.Errorf("map element length is not equal to grouped rows. Query = %v and grouped = %v (RowId = %v)", rowSize, groupedRowSize, rowId)
		}
	}
}

type EchoContextMock struct {
	data map[string]string
}

func (e *EchoContextMock) QueryParam(name string) string {
	v, ok := e.data[name]
	if !ok {
		return ""
	}
	return v
}

// go test -v -run ^TestMHSExecutorDictionary -timeout 0 ./mhs
func TestMHSExecutorDictionary(t *testing.T) {
	mhsExecutorDictionary(t)
}

func mhsExecutorDictionary(t *testing.T) {
	stateSingleton := state.GetStateSingleton()
	defer stateSingleton.Pool.Close()

	mhsExecutor := NewExecutor(stateSingleton.Pool)
	mhsExecutor.SetBehaviorOnError(false)

	fileName := mhsExecutor.getTemporaryFilePath("test-input")
	log.Println("MHS test input file name: ", fileName)

	file, err := os.Create(fileName)
	defer func() {
		err := file.Close()
		if err != nil {
			t.Errorf("error closing file: %+v", err)
		}

		err = os.Remove(fileName)
		if err != nil {
			t.Errorf("error removing file: %+v", err)
		}
	}()

	if err != nil {
		t.Errorf("can't create a file %v", err)
	}

	cMock := EchoContextMock{data: map[string]string{
		"dictionary": "501",
	}}
	qh, err := NewQueryHelper(&cMock)
	if err != nil {
		t.Errorf("can't create a query helper %v", err)
	}

	log.Println(qh.CreateQuery())
	err = mhsExecutor.GetSentences(context.Background(), file, *qh, 1000)
	if err != nil {
		t.Errorf("error executing mhs %v", err)
	}

	fileStats, err := os.Stat(fileName)
	if err != nil {
		t.Errorf("error while checking file size %v", err)
	}
	if fileStats.Size() < int64(10) {
		t.Error("output file is empty")
	}
}

func getValueToAdd(v int) (int, error) {
	strVal := strconv.Itoa(v)
	newVal := "1" + strVal[1:]
	return strconv.Atoi(newVal)
}

// go test -v -run ^TestOptimalMHSValues -timeout 0 ./mhs
func TestOptimalMHSValues(t *testing.T) {

	log.SetOutput(io.Discard)

	var slice, sliceStep int
	var sentencePV, sentenceStep int

	for slice = 50; slice <= 100_000; {
		for sentencePV = 10; sentencePV <= 100_000; {
			testName := fmt.Sprintf("Test_Slice_%d_Sentence_%v", slice, sentencePV)

			t.Run(testName, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("ERROR: %s\n", testName)
					}
				}()

				t.Setenv("MHS_MAX_SLICE_SIZE", fmt.Sprintf("%v", slice))
				t.Setenv("MHS_MAX_SENTENCES_PER_VALUE", fmt.Sprintf("%v", sentencePV))

				mhsExecutorDictionary(t)
			})

			runtime.GC()

			sentenceStep, _ = getValueToAdd(sentencePV)
			sentencePV += sentenceStep
		}
		sliceStep, _ = getValueToAdd(slice)
		slice += sliceStep
	}
}
