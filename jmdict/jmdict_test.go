package jmdict

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEntryMarshaling(t *testing.T) {
	var xmlData = []byte(`
<entry>
<ent_seq>1000220</ent_seq>
<k_ele>
<keb>明白</keb>
<ke_pri>ichi1</ke_pri>
<ke_pri>news1</ke_pri>
<ke_pri>nf10</ke_pri>
</k_ele>
<r_ele>
<reb>めいはく</reb>
<re_pri>ichi1</re_pri>
<re_pri>news1</re_pri>
<re_pri>nf10</re_pri>
</r_ele>
<r_ele>
<reb>TEST</reb>
</r_ele>
</entry>
`)

	expected := Entry{
		EntryId: 1000220,
		Readings: []Reading{
			{Word: "明白", Dictionary: []string{"ichi1", "news1", "nf10"}, IsKanji: true, OrderId: 1000220<<2 + 1},
			{Word: "めいはく", Dictionary: []string{"ichi1", "news1", "nf10"}, IsKanji: false, OrderId: 1000220<<2 + 2},
			{Word: "TEST", Dictionary: []string{}, IsKanji: false, OrderId: 1000220<<2 + 3},
		},
	}
	var test Entry
	err := xml.Unmarshal(xmlData, &test)
	if err != nil {
		t.Error(err)
	} else {
		diff := cmp.Diff(test, expected)
		if diff != "" {
			t.Errorf("%+v", diff)
		}
	}

}
