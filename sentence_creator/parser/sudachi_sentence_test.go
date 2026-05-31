package parser

import (
	"bytes"
	"testing"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
	"github.com/google/go-cmp/cmp"
)

func TestCreateSudachiTSVParsingChan(t *testing.T) {
	EntryFile := []byte(`
1297 	 jpn 	 きみ に ちょっと し た もの を もっ て き た よ 。
4702 	 jpn 	 何 か し て み ましょう 。
4703 	 jpn 	 私 は 眠ら なけれ ば なり ませ ん 。
4704 	 jpn 	 何 し てる の ？
4705 	 jpn 	 今日 は ６ 月 １８ 日 で 、 ムーリエル の 誕生日 です ！
`)
	expected := []*tatoeba.SentenceTokens{
		&tatoeba.SentenceTokens{Id: 1297, Tokens: &[]string{"きみ", "に", "ちょっと", "し", "た", "もの", "を", "もっ", "て", "き", "た", "よ", "。"}},
		&tatoeba.SentenceTokens{Id: 4702, Tokens: &[]string{"何", "か", "し", "て", "み", "ましょう", "。"}},
		&tatoeba.SentenceTokens{Id: 4703, Tokens: &[]string{"私", "は", "眠ら", "なけれ", "ば", "なり", "ませ", "ん", "。"}},
		&tatoeba.SentenceTokens{Id: 4704, Tokens: &[]string{"何", "し", "てる", "の", "？"}},
		&tatoeba.SentenceTokens{Id: 4705, Tokens: &[]string{"今日", "は", "６", "月", "１８", "日", "で", "、", "ムーリエル", "の", "誕生日", "です", "！"}},
	}

	buffer := bytes.NewBuffer(EntryFile)
	sChan := CreateSudachiTSVParsingChan(buffer, 1)

	output := make([]*tatoeba.SentenceTokens, 0)

	for s := range sChan {
		output = append(output, s)
	}
	diff := cmp.Diff(output, expected)
	if diff != "" {
		t.Errorf("%+v", diff)
	}
}
