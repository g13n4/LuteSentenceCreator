package domain

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/kanji"
	"github.com/g13n4/LuteSentencePicker/mhs"
	"github.com/g13n4/LuteSentencePicker/tatoeba"
	"github.com/g13n4/LuteSentencePicker/utils"
)

type KanjiRepository interface {
	Save(ctx context.Context, obj *kanji.Kanji) error
	BulkSave(kanjis []*kanji.Kanji) error
	GetUniqueFields(ctx context.Context, field string) ([]int, error)
}

type DictionaryRepository interface {
	BulkSave(dictionaries *[]*jmdict.Dictionary) error
	GetDictionaries(ctx context.Context) ([]*jmdict.Dictionary, error)
}

type DictionaryCategoryRepository interface {
	BulkSave(dictionaries *[]*jmdict.DictionaryCategory) error
}

type EntryRepository interface {
	ToCache(key string, value utils.PostgresID)
	FromCache(name string) (utils.PostgresID, bool)

	GetIdByKanjiReading(ctx context.Context, kanjiReading string) (utils.PostgresID, error)
	GetIdByReading(ctx context.Context, reading string) (utils.PostgresID, error)
	BulkSave(objs []*jmdict.Entry) error
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

type MHSRepository interface {
	GetSentences(ctx context.Context, mshq *mhs.QueryHelper, permuts, limit int) ([]string, error)
}
