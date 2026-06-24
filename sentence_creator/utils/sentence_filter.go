package utils

type SentenceFilter struct {
	sentenceFitStatus    map[int]bool
	sentenceReadingCount map[int]int
	maximumReadingAmount int
}

func NewSentenceFilter() *SentenceFilter {
	maxReadingAmount := GetEnvIntValue("MAXIMUM_READING_AMOUNT", 50)

	return &SentenceFilter{
		sentenceFitStatus:    make(map[int]bool, 100_000),
		sentenceReadingCount: make(map[int]int, 10_000),
		maximumReadingAmount: maxReadingAmount,
	}
}

// Fits checks if a sentence whose reading ids were provided should be included in a database or not
func (sf *SentenceFilter) Fits(sentenceId int, sentenceReading *[]*int) bool {
	var fitStatus bool
	if sf.fitsInFilter(sentenceReading) {
		sf.updateFilter(sentenceReading)
		fitStatus = true
	}

	sf.sentenceFitStatus[sentenceId] = fitStatus
	return fitStatus
}

func (sf *SentenceFilter) updateFilter(sentenceReading *[]*int) {
	for _, v := range *sentenceReading {
		if v != nil {
			sf.sentenceReadingCount[*v]++
		}
	}
}

func (sf *SentenceFilter) fitsInFilter(sentenceReading *[]*int) bool {
	for _, v := range *sentenceReading {
		if v != nil {
			if sf.sentenceReadingCount[*v] < sf.maximumReadingAmount {
				return true
			}
		}
	}
	return false
}

func (sf *SentenceFilter) getSentenceIdByFitStatus(status bool) *[]int {
	output := make([]int, 0, len(sf.sentenceFitStatus))
	for k, v := range sf.sentenceFitStatus {
		if v == status {
			output = append(output, k)
		}
	}
	return &output
}

func (sf *SentenceFilter) GetFitSentenceId() *[]int {
	return sf.getSentenceIdByFitStatus(true)
}

func (sf *SentenceFilter) GetUnfitSentenceId() *[]int {
	return sf.getSentenceIdByFitStatus(false)
}
