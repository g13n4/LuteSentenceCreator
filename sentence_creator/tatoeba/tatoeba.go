package tatoeba

import (
	"fmt"
	"strings"
)

type Sentence struct {
	Id         int
	Text       string
	isFiltered bool
}

func (s *Sentence) String() string {
	return fmt.Sprintf("%d: %s", s.Id, s.Text)
}

type SentenceTokens struct {
	Id     int
	Tokens *[]string
}

func (st *SentenceTokens) String() string {
	return fmt.Sprintf("%d: %s", st.Id, strings.Join(*st.Tokens, ","))
}
