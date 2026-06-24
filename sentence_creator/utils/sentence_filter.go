package utils

type SentenceFilter struct {
	sentenceReadingCount map[int]int
	maximumReadingAmount int
}

func NewSentenceFilter() *SentenceFilter {
	maxReadingAmount := GetEnvIntValue("MAXIMUM_READING_AMOUNT", 50)

	return &SentenceFilter{
		sentenceReadingCount: make(map[int]int, 10_000),
		maximumReadingAmount: maxReadingAmount,
	}
}

// Fits checks if a sentence whose reading ids were provided should be included in a database or not
func (sf *SentenceFilter) Fits(sentenceReading *[]*int) bool {
	if sf.fitsInFilter(sentenceReading) {
		sf.addToFilter(sentenceReading)
		return true
	}
	return false
}

func (sf *SentenceFilter) addToFilter(sentenceReading *[]*int) {
	for _, v := range *sentenceReading {
		if v != nil {
			sf.sentenceReadingCount[*v]++
		}
	}
}

func (sf *SentenceFilter) fitsInFilter(sentenceReading *[]*int) bool {
	for _, v := range *sentenceReading {
		if v != nil {
			if sf.sentenceReadingCount[*v] > sf.maximumReadingAmount {
				return false
			}
		}
	}
	return true
}
