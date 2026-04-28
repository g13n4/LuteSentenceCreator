package domain

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/kanji"
	"github.com/g13n4/LuteSentencePicker/tatoeba"
	"github.com/g13n4/LuteSentencePicker/utils"
)

type KanjiRepository interface {
	Save(ctx context.Context, kanji kanji.Kanji) error
	BulkSave(ctx context.Context, kanjis []kanji.Kanji) error
}

type DictionaryRepository interface {
	GetIdByName(ctx context.Context, name string) (utils.PostgresID, error)
	Save(ctx context.Context, obj jmdict.DictionaryEntry) error
}

type EntryRepository interface {
	GetIdByKanjiReading(ctx context.Context, kanjiReading string) (utils.PostgresID, error)
	GetIdByReading(ctx context.Context, reading string) (utils.PostgresID, error)
	Save(ctx context.Context, obj jmdict.Entry) error
	BulkSave(ctx context.Context, objs []jmdict.Entry) error
}

type SentenceRepository interface {
	Save(ctx context.Context, sentence tatoeba.Sentence) error
	BulkSave(ctx context.Context, sentences []tatoeba.Sentence) error
}

type ConnectionsRepository interface {
	SaveDictionaryEntry(ctx context.Context, dictionaryEntry jmdict.DictionaryEntry) error
}
