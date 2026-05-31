package parser

import (
	"bytes"
	"testing"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
	"github.com/google/go-cmp/cmp"
)

func TestCreateTSVParsingChan(t *testing.T) {
	EntryFile := []byte(`
1297	jpn	きみにちょっとしたものをもってきたよ。
4702	jpn	何かしてみましょう。
4703	jpn	私は眠らなければなりません。
4704	jpn	何してるの？
4705	jpn	今日は６月１８日で、ムーリエルの誕生日です！
4706	jpn	お誕生日おめでとうムーリエル！
4707	jpn	ムーリエルは２０歳になりました。
4708	jpn	パスワードは「Muiriel」です。
4709	jpn	すぐに戻ります。
`)
	expected := []*tatoeba.Sentence{
		&tatoeba.Sentence{Id: 1297, Text: "きみにちょっとしたものをもってきたよ。"},
		&tatoeba.Sentence{Id: 4702, Text: "何かしてみましょう。"},
		&tatoeba.Sentence{Id: 4703, Text: "私は眠らなければなりません。"},
		&tatoeba.Sentence{Id: 4704, Text: "何してるの？"},
		&tatoeba.Sentence{Id: 4705, Text: "今日は６月１８日で、ムーリエルの誕生日です！"},
		&tatoeba.Sentence{Id: 4706, Text: "お誕生日おめでとうムーリエル！"},
		&tatoeba.Sentence{Id: 4707, Text: "ムーリエルは２０歳になりました。"},
		&tatoeba.Sentence{Id: 4708, Text: "パスワードは「Muiriel」です。"},
		&tatoeba.Sentence{Id: 4709, Text: "すぐに戻ります。"},
	}

	buffer := bytes.NewBuffer(EntryFile)
	sChan := CreateTSVParsingChan(buffer, 1)

	output := make([]*tatoeba.Sentence, 0)

	for s := range sChan {
		output = append(output, s)
	}
	diff := cmp.Diff(output, expected)
	if diff != "" {
		t.Errorf("%+v", diff)
	}
}
