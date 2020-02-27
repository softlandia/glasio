//(c) softland 2019
//softlandia@gmail.com

package glasio

import (
	"testing"
)

func TestLoadStdMnemonicDic(t *testing.T) {
	_, err := LoadStdMnemonicDic() //file ini not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 1 return error == nil\n")
	}
	_, err = LoadStdMnemonicDic("data\\mnemonic.ini") //file ini exist, return error == nil
	if err != nil {
		t.Errorf("<LoadStdMnemonicDic> on test 2 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdMnemonicDic("data\\mn.ini") //file ini not exist, return error
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 3 expect error != nil, return nil\n")
	}
	_, err = LoadStdMnemonicDic("data\\dic.ini") //file ini exist, section not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 4 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdMnemonicDic("data\\dic0.ini") //file ini exist, file empty, return error != nil
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 5 expect error = nil, return error: %s\n", err)
	}
}

func TestLoadStdVocabularyDictionary(t *testing.T) {
	_, err := LoadStdVocabularyDictionary() //file ini not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 1 return error == nil\n")
	}
	_, err = LoadStdVocabularyDictionary("data\\dic.ini") //file ini exist, return error == nil
	if err != nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 2 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdVocabularyDictionary("data\\mn.ini") //file ini not exist, return error
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 3 expect error != nil, return nil\n")
	}
	_, err = LoadStdVocabularyDictionary("data\\mnemonic.ini") //file ini exist, section not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 4 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdVocabularyDictionary("data\\mnemonic0.ini") //file ini exist, file empty, return error != nil
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 5 expect error = nil, return error: %s\n", err)
	}
}
