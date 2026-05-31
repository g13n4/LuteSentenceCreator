package application

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/utils"
)

type Button struct {
	Value string
}

type IndexButtons struct {
	Grades      []Button
	StrokeCount []Button

	News []Button
	Ichi []Button
	Spec []Button
	Gai  []Button
	Nf   []Button
}

type FrontendData struct {
	kanjiRepo domain.KanjiRepository

	dictRepo domain.DictionaryRepository
}

func (bf *FrontendData) ToButton(values []int) []Button {
	out := make([]Button, len(values))
	for idx, v := range values {
		out[idx] = Button{Value: utils.IntegerToSafeString(v)}
	}
	return out
}

func (bf *FrontendData) GetIndexButtons() (*IndexButtons, error) {
	var ib IndexButtons
	grades, err := bf.kanjiRepo.GetUniqueFields(context.Background(), "grade")
	if err != nil {
		return nil, err
	}

	ib.Grades = bf.ToButton(grades)

	sCount, err := bf.kanjiRepo.GetUniqueFields(context.Background(), "stroke_count")
	if err != nil {
		return nil, err
	}

	ib.StrokeCount = bf.ToButton(sCount)

	dictionaries, err := bf.dictRepo.GetDictionaries(context.Background())
	if err != nil {
		return nil, err
	}

	ib.News = make([]Button, 0)
	ib.Ichi = make([]Button, 0)
	ib.Spec = make([]Button, 0)
	ib.Gai = make([]Button, 0)
	ib.Nf = make([]Button, 0)
	for _, d := range dictionaries {
		if d.Category == jmdict.DictionaryNewsValue {
			ib.News = append(ib.News, Button{Value: utils.IntegerToSafeString(d.Number)})
		}
		if d.Category == jmdict.DictionaryIchiValue {
			ib.Ichi = append(ib.Ichi, Button{Value: utils.IntegerToSafeString(d.Number)})
		}
		if d.Category == jmdict.DictionarySpecValue {
			ib.Spec = append(ib.Spec, Button{Value: utils.IntegerToSafeString(d.Number)})
		}
		if d.Category == jmdict.DictionaryGaiValue {
			ib.Gai = append(ib.Gai, Button{Value: utils.IntegerToSafeString(d.Number)})
		}
		if d.Category == jmdict.DictionaryNFValue {
			ib.Nf = append(ib.Nf, Button{Value: utils.IntegerToSafeString(d.Number)})
		}
	}
	return &ib, nil
}

func NewButtonsFrontend(db repository.DBSaver) *FrontendData {
	kr := repository.NewKanjiRepository(db)
	dr := repository.NewDictionaryRepository(db)

	return &FrontendData{kanjiRepo: kr, dictRepo: dr}
}
