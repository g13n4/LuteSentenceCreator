package jmdict

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDictionaryPool1(t *testing.T) {
	dp := NewDictionaryPool()

	dictValue := "ichi1"
	dictObj := Dictionary{Id: DictionaryIchiValue*100 + 1, Name: dictValue, Category: DictionaryIchiValue, Number: 1}
	dictCategoryName := dp.GetCategoryName(dictValue)
	dictCategoryObj := dp.GetCategory(dictCategoryName)

	testDictObj, testDictCatObj := dp.GetDictionaryData(dictValue)

	diff := cmp.Diff(*testDictObj, dictObj)
	if diff != "" {
		t.Errorf("%+v", diff)
	}

	diff = cmp.Diff(dictCategoryObj, testDictCatObj)
	if diff != "" {
		t.Errorf("%+v", diff)
	}

	// check if the dictionary objects are not created again but are from cache
	testDictObj1, testDictCatObj1 := dp.GetDictionaryData("ichi1")
	testDictObj2, testDictCatObj2 := dp.GetDictionaryData("ichi1")
	diff = cmp.Diff(&testDictObj1, &testDictObj2)
	if diff != "" {
		t.Errorf("%+v", diff)
	}

	diff = cmp.Diff(&testDictCatObj1, &testDictCatObj2)
	if diff != "" {
		t.Errorf("%+v", diff)
	}
}

func TestDictionaryPool2(t *testing.T) {
	dp := NewDictionaryPool()

	dictValue := "nf41"
	dictObj := Dictionary{Id: DictionaryNFValue*100 + 41, Name: dictValue, Category: DictionaryNFValue, Number: 41}
	dictCategoryName := dp.GetCategoryName(dictValue)
	dictCategoryObj := dp.GetCategory(dictCategoryName)

	testDictObj, testDictCatObj := dp.GetDictionaryData(dictValue)

	diff := cmp.Diff(*testDictObj, dictObj)
	if diff != "" {
		t.Errorf("%+v", diff)
	}

	diff = cmp.Diff(dictCategoryObj, testDictCatObj)
	if diff != "" {
		t.Errorf("%+v", diff)
	}
}
