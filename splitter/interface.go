package splitter

type SentenceIn struct {
	Id   int
	Text string
}

type SentenceOut struct {
	Id     int
	Tokens *[]string
}
