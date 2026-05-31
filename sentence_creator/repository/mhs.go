package repository

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

	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/mhs"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
	"github.com/jackc/pgx/v5"
)

type mhsRepository struct {
	db DBSaver
}

func NewMHSRepository(db DBSaver) domain.MHSRepository {
	return &mhsRepository{db: db}
}

func (mhsr *mhsRepository) GetSentences(ctx context.Context, outputFile *os.File, mshq mhs.QueryHelper, permuts, limit int) error {
	var sChan <-chan *[]int
	var err error

	if !mshq.IsKanji() {
		sChan, err = mhsr.getSentencesForEntry(ctx, mshq)
	} else {
		sChan, err = mhsr.getSentencesForKanji(ctx, mshq)
	}
	if err != nil {
		return err
	}
	fileIn, fileOut, err := mhsr.createSentenceSetFileInput(sChan)
	if err != nil {
		return err
	}
	sentenceIds, err := mhsr.processSentenceSetFile(fileIn, fileOut, permuts)
	if err != nil {
		return err
	}

	return mhsr.saveSentencesSet(ctx, outputFile, sentenceIds, limit)
}

func (mhsr *mhsRepository) getSentencesForEntry(ctx context.Context, mshq mhs.QueryHelper) (<-chan *[]int, error) {
	query := "SELECT smr.r_id, smr.s_id from sentences__mtm__readings smr JOIN readings r ON smr.r_id = r.id JOIN dictionaries__mtm__entries dme ON r.entry = dme.entry WHERE "
	rows, err := mhsr.db.Query(ctx, query+mshq.GetSQLCondition("dme"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return mhsr.createSentenceSetChannel(rows)
}

func (mhsr *mhsRepository) getSentencesForKanji(ctx context.Context, mshq mhs.QueryHelper) (<-chan *[]int, error) {
	query := "SELECT smr.r_id, smr.s_id from sentences__mtm__readings smr JOIN readings__mtm__kanjis rmk ON rmk.r_id = smr.r_id JOIN kanjis k ON rmk.k_id = k.id WHERE "
	rows, err := mhsr.db.Query(ctx, query+mshq.GetSQLCondition("k"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return mhsr.createSentenceSetChannel(rows)
}

func (mhsr *mhsRepository) createSentenceSetChannel(rows pgx.Rows) (<-chan *[]int, error) {
	out := make(chan *[]int)
	rToS := make(map[int][]int)

	for rows.Next() {
		var readingId int
		var sentenceId int

		err := rows.Scan(&readingId, &sentenceId)
		if err != nil {
			return nil, err
		}

		_, ok := rToS[readingId]
		if !ok {
			rToS[readingId] = make([]int, 0)
		}
		rToS[readingId] = append(rToS[readingId], sentenceId)

	}
	go func() {
		for _, v := range rToS {
			out <- &v
		}
	}()

	return out, nil
}

func (mhsr *mhsRepository) createSentenceSetFileInput(senChan <-chan *[]int) (string, string, error) {
	tempPrefix := fmt.Sprintf("%v", time.Now().Unix())
	tempNameInput := fmt.Sprintf("input-%v.dat", tempPrefix)
	tempNameOutput := fmt.Sprintf("output-%v.dat", tempPrefix)

	mhsInput := filepath.Join(os.TempDir(), tempNameInput)
	mhsOutput := filepath.Join(os.TempDir(), tempNameOutput)

	file, err := os.Create(mhsInput)
	defer func() {
		err := file.Close()
		panic(err)
	}()

	if err != nil {
		return "", "", err
	}

	writer := bufio.NewWriter(file)
	for line := range senChan {
		lineLen := len(*line)
		for idx, num := range *line {
			_, err := writer.WriteString(strconv.Itoa(num))
			if err != nil {
				return "", "", err
			}
			if lineLen-1 == idx {
				_, err = writer.WriteString(strconv.Itoa(num) + "\n")
			} else {
				_, err = writer.WriteString(strconv.Itoa(num) + " ")
			}
			if err != nil {
				return "", "", err
			}
		}
	}
	err = writer.Flush()
	if err != nil {
		return "", "", err
	}

	return mhsInput, mhsOutput, nil
}

func (mhsr *mhsRepository) processSentenceSetFile(input, output string, permuts int) (<-chan *string, error) {
	mhsExec := os.Getenv("MHS_PATH")
	if mhsExec == "" {
		return nil, errors.New("MHS_PATH environment variable is not set")
	}

	cmd := exec.Command(mhsExec, input, output, "-a pmmcs")
	outputChan := make(chan *string)

	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Command failed: %s\nOutput: %s", err, output)
	}

	file, closer, err := utils.OpenFile(output)
	defer func() {
		err := closer()
		if err != nil {
			panic(err)
		}
	}()

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
			numbers := strings.Split(line, " ")
			for _, num := range numbers {
				_, ok := sentenceSet[num]
				if ok {
					continue
				}

				outputChan <- &num
				sentenceSet[num] = struct{}{}
			}
			lineCounter++

			if lineCounter >= permuts {
				return
			}
		}
	}()

	return outputChan, err
}

func (mhsr *mhsRepository) saveSentencesSet(ctx context.Context, outputFile *os.File, sentenceIdChan <-chan *string, limit int) error {
	ids := make([]string, 0)

	for sentenceId := range sentenceIdChan {
		ids = append(ids, *sentenceId)
	}

	query := fmt.Sprintf("SELECT sentence from sentences WHERE id = ANY($1) LIMIT %v", limit)
	rows, err := mhsr.db.Query(ctx, query, ids)

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(outputFile)
	for rows.Next() {
		var sentence string

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

func (mhsr *mhsRepository) collectSentenceSet(ctx context.Context, sentenceIdChan <-chan *string, limit int) ([]string, error) {
	ids := make([]string, 0)

	for sentenceId := range sentenceIdChan {
		ids = append(ids, *sentenceId)
	}

	query := fmt.Sprintf("SELECT sentence from sentences WHERE id = ANY($1) LIMIT %v", limit)
	rows, err := mhsr.db.Query(ctx, query, ids)

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
