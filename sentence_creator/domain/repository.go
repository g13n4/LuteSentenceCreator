package domain

import (
	"context"

	jmdict2 "github.com/g13n4/LuteSentencePicker/sentence_creator/jmdict"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/kanji"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

type KanjiRepository interface {
	Save(ctx context.Context, obj *kanji.Kanji) error
	BulkSave(kanjis []*kanji.Kanji) error
	GetUniqueFields(ctx context.Context, field string) ([]int, error)
}

type DictionaryRepository interface {
	BulkSave(dictionaries *[]*jmdict2.Dictionary) error
	GetDictionaries(ctx context.Context) ([]*jmdict2.Dictionary, error)
}

type DictionaryCategoryRepository interface {
	BulkSave(dictionaries *[]*jmdict2.DictionaryCategory) error
}

type EntryRepository interface {
	ToCache(key string, value utils.PostgresID)
	FromCache(name string) (utils.PostgresID, bool)

	GetIdByKanjiReading(ctx context.Context, kanjiReading string) (utils.PostgresID, error)
	GetIdByReading(ctx context.Context, reading string) (utils.PostgresID, error)
	BulkSave(objs []*jmdict2.Entry) error
}

type SentenceRepository interface {
	Save(ctx context.Context, sentence tatoeba.Sentence) error
	BulkSave(sentences []*tatoeba.Sentence) error
}

type ConnectionsRepository[T any] interface {
	BulkSave(objs []T) error
}

type DBStateRepository interface {
	SetStatus(ctx context.Context, val int) error
	GetStatus(ctx context.Context) (int, error)
}

type ExecutorSentenceRepository interface {
	GetSentences()
}
