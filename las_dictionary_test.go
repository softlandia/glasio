//(c) softland 2019
//softlandia@gmail.com

package glasio

import (
	"path/filepath"
	"testing"
)

func TestLoadStdMnemonicDic(t *testing.T) {
	_, err := LoadStdMnemonicDic() //file ini not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 1 return error == nil\n")
	}
	_, err = LoadStdMnemonicDic(filepath.Join("data", "mnemonic.ini")) //file ini exist, return error == nil
	if err != nil {
		t.Errorf("<LoadStdMnemonicDic> on test 2 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdMnemonicDic(filepath.Join("data", "mn.ini")) //file ini not exist, return error
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 3 expect error != nil, return nil\n")
	}
	_, err = LoadStdMnemonicDic(filepath.Join("data", "dic.ini")) //file ini exist, section not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 4 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdMnemonicDic(filepath.Join("data", "dic0.ini")) //file ini exist, file empty, return error != nil
	if err == nil {
		t.Errorf("<LoadStdMnemonicDic> on test 5 expect error = nil, return error: %s\n", err)
	}
}

func TestLoadStdVocabularyDictionary(t *testing.T) {
	_, err := LoadStdVocabularyDictionary() //file ini not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 1 return error == nil\n")
	}
	_, err = LoadStdVocabularyDictionary(filepath.Join("data", "dic.ini")) //file ini exist, return error == nil
	if err != nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 2 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdVocabularyDictionary(filepath.Join("data", "mn.ini")) //file ini not exist, return error
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 3 expect error != nil, return nil\n")
	}
	_, err = LoadStdVocabularyDictionary(filepath.Join("data", "mnemonic.ini")) //file ini exist, section not exist, return error != nil
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 4 expect error = nil, return error: %s\n", err)
	}
	_, err = LoadStdVocabularyDictionary(filepath.Join("data", "mnemonic0.ini")) //file ini exist, file empty, return error != nil
	if err == nil {
		t.Errorf("<LoadStdVocabularyDictionary> on test 5 expect error = nil, return error: %s\n", err)
	}
}
