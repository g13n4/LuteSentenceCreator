package connections

type DictionaryEntry struct {
	Entry                int
	DictionaryCategoryId int
	DictionaryId         int
}

type ReadingKanji struct {
	ReadingId int
	KanjiId   int
}

type SentenceReading struct {
	SentenceId int
	ReadingId  int
}
