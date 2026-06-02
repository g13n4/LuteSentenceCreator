package parser

import (
	"bytes"
	"testing"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/jmdict"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/kanji"
	"github.com/google/go-cmp/cmp"
)

func toPtr(i int) *int {
	return &i
}

func TestCreateXMLParsingChanEntry(t *testing.T) {
	EntryFile := []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<JMdict>
<entry>
<ent_seq>1244840</ent_seq>
<k_ele>
<keb>駆る</keb>
<ke_pri>news2</ke_pri>
<ke_pri>nf29</ke_pri>
<ke_pri>spec2</ke_pri>
</k_ele>
<k_ele>
<keb>駈る</keb>
</k_ele>
<r_ele>
<reb>かる</reb>
<re_pri>news2</re_pri>
<re_pri>nf29</re_pri>
<re_pri>spec2</re_pri>
</r_ele>
</entry>
<entry>
<ent_seq>1552120</ent_seq>
<k_ele>
<keb>流す</keb>
<ke_pri>ichi1</ke_pri>
<ke_pri>news1</ke_pri>
<ke_pri>nf08</ke_pri>
</k_ele>
<r_ele>
<reb>ながす</reb>
<re_pri>ichi1</re_pri>
<re_pri>news1</re_pri>
<re_pri>nf08</re_pri>
</r_ele>
</entry>
</JMdict>
`)
	expected := []*jmdict.Entry{
		&jmdict.Entry{EntryId: 1244840, Readings: []jmdict.Reading{
			jmdict.Reading{Word: "駆る", Dictionary: []string{"news2", "nf29", "spec2"}, IsKanji: true, CombinedId: 1244840*100 + 1},
			jmdict.Reading{Word: "駈る", Dictionary: []string{}, IsKanji: true, CombinedId: 1244840*100 + 2},
			jmdict.Reading{Word: "かる", Dictionary: []string{"news2", "nf29", "spec2"}, IsKanji: false, CombinedId: 1244840*100 + 3},
		}},
		&jmdict.Entry{EntryId: 1552120, Readings: []jmdict.Reading{
			jmdict.Reading{Word: "流す", Dictionary: []string{"ichi1", "news1", "nf08"}, IsKanji: true, CombinedId: 1552120*100 + 1},
			jmdict.Reading{Word: "ながす", Dictionary: []string{"ichi1", "news1", "nf08"}, IsKanji: false, CombinedId: 1552120*100 + 2},
		}},
	}

	buffer := bytes.NewBuffer(EntryFile)
	eChan := CreateXMLParsingChan[*jmdict.Entry](buffer, jmdict.EntryNodeName, 1)

	output := make([]*jmdict.Entry, 0)

	for k := range eChan {
		output = append(output, k)
	}
	diff := cmp.Diff(output, expected)
	if diff != "" {
		t.Errorf("%+v", diff)
	}
}

func TestCreateXMLParsingChanKanji(t *testing.T) {
	KanjiFile := []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<kanjidic2>
<header>
<!-- KANJIDIC 2 - XML format kanji database combining the KANJIDIC
	and KANJD212 files plus the kanji from JIS X 0213.
-->
<file_version>4</file_version>
<database_version>2026-109</database_version>
<date_of_creation>2026-04-19</date_of_creation>
</header>
<character>
<literal>亜</literal>
<misc>
<grade>8</grade>
<stroke_count>7</stroke_count>
<freq>1509</freq>
<jlpt>1</jlpt>
</misc>
</character>
<character>
<literal>訊</literal>
<misc>
<grade>9</grade>
<stroke_count>10</stroke_count>
</misc>
</character>
<character>
<literal>旵</literal>
<misc>
<stroke_count>7</stroke_count>
</misc>
</character>
</kanjidic2>
`)
	expected := []*kanji.Kanji{
		&kanji.Kanji{Literal: "亜", Grade: toPtr(8), StrokeCount: toPtr(7), Frequency: toPtr(1509), JLPT: toPtr(1)},
		&kanji.Kanji{Literal: "訊", Grade: toPtr(9), StrokeCount: toPtr(10), Frequency: nil, JLPT: nil},
		&kanji.Kanji{Literal: "旵", Grade: nil, StrokeCount: toPtr(7), Frequency: nil, JLPT: nil},
	}

	buffer := bytes.NewBuffer(KanjiFile)
	kChan := CreateXMLParsingChan[*kanji.Kanji](buffer, kanji.NodeName, 1)
	output := make([]*kanji.Kanji, 0)

	for k := range kChan {
		output = append(output, k)
	}
	diff := cmp.Diff(output, expected)
	if diff != "" {
		t.Errorf("%+v", diff)
	}
}
