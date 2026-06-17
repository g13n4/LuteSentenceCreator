package mhs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

	err := mhsExecutor.createMHSFileInput(mhsInput, testMap)
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
		err = mhsExecutor.processSentenceSetFile(mhsInput, mhsOutput, testChan)
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

	for line := range testChan {
		count++
		output += *line
	}

	if count < 3 {
		t.Errorf("expected at least 3 numbers in a set and got %v", count)
	}
	if output == "" {
		t.Errorf("output is empty")
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

// go test -v -run ^TestMHSExecutorKanji ./mhs
func TestMHSExecutorKanji(t *testing.T) {
	stateSingleton := state.GetStateSingleton()

	mhsExecutor := NewExecutor(stateSingleton.Pool)

	mhsInput := filepath.Join(os.TempDir(), "test-intput.dat")
	file, err := os.Create(mhsInput)
	defer func() {
		err := file.Close()
		if err != nil {
			t.Errorf("error closing file: %+v", err)
		}

		err = os.Remove(mhsInput)
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

	err = mhsExecutor.GetSentences(context.Background(), file, *qh, 1000)
	if err != nil {
		t.Errorf("error executing mhs %v", err)

	}
}
