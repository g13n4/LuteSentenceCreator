package tatoeba

import (
	"fmt"
)

type Sentence struct {
	Id   int
	Text string
}

func (s *Sentence) String() string {
	return fmt.Sprintf("%d: %s", s.Id, s.Text)
}
