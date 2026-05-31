package splitter

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func CreateSentenceToEntryParsingChan(sentenceChan <-chan *SentenceIn) <-chan *SentenceOut {
	cmd := exec.Command("./sudachi-exec -w")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	output := make(chan *SentenceOut)

	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var out SentenceOut
			tokens := strings.Split(scanner.Text(), " ")
			out.Id, err = strconv.Atoi(tokens[0])
			if err != nil {
				panic("something wrong with the sentence id embedding")
			}
			out.Tokens = &tokens
			output <- &out
		}
		defer close(output)
	}()

	go func() {
		select {
		case <-done:
			return
		case s := <-sentenceChan:
			_, err := fmt.Fprint(stdin, s.Id, "\\", s.Text)
			if err != nil {
				panic(err)
			}
			return
		}
	}()

	err = stdin.Close()
	if err != nil {
		panic(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	return output
}
