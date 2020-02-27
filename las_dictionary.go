package glasio

import (
	"fmt"

	"gopkg.in/ini.v1"
)

const lasStdMnemonicIniFileName = "mnemonic.ini"
const lasStdMnemonicIniSection = "mnemonic"

// TMnemonic - dictionary of std log name == mnemonics
type TMnemonic = map[string]string

//LoadStdMnemonicDic - return std mnemonic map
//read ini file, fill dictionary
//return empty map if occure error on reading ini file with dictionary, may be add after
func LoadStdMnemonicDic(fileName ...string) (TMnemonic, error) {
	if len(fileName) == 0 {
		fileName = make([]string, 1)
		fileName[0] = lasStdMnemonicIniFileName
	}
	mnemonic := make(TMnemonic)
	iniMnemonic, err := ini.Load(fileName[0])
	if err != nil {
		return mnemonic, fmt.Errorf("<GetStdMnemonic> can't read ini file '%s'. error: %v", fileName[0], err)
	}
	sec, err := iniMnemonic.GetSection(lasStdMnemonicIniSection)
	if err != nil {
		return mnemonic, fmt.Errorf("<GetStdMnemonic> can't read section 'mnemonic' from ini file '%s'. error: %v", fileName[0], err)
	}
	for _, s := range sec.KeyStrings() {
		mnemonic[s] = sec.Key(s).Value() //добавление в словарь значения 'sec.Key(s).Value()' с ключём 's'
	}
	return mnemonic, nil
}

//TVocDic - dictionary of std log name == mnemonics
type TVocDic = map[string]string

const lasStdVocDicFileName = "dic.ini"
const lasStdVocDicSection = "LOG"

//LoadStdVocabularyDictionary - return lookup map
//read ini file, fill dictionary
//return empty map if occure error on reading ini file with dictionary, may be add after
func LoadStdVocabularyDictionary(fileName ...string) (TVocDic, error) {
	if len(fileName) == 0 {
		fileName = make([]string, 1)
		fileName[0] = lasStdVocDicFileName
	}

	//create empty map
	vocDic := make(TVocDic)

	iniVocDic, err := ini.Load(fileName[0])
	if err != nil {
		return vocDic, fmt.Errorf("<GetStdVocabularyDictionary> can't read ini file '%s'. error: %v", fileName[0], err)
	}
	sec, err := iniVocDic.GetSection(lasStdVocDicSection)
	if err != nil {
		return vocDic, fmt.Errorf("<GetStdVocabularyDictionary> can't read section 'LOG' from ini file '%s'. error: %v", fileName[0], err)
	}

	//fill dictionary
	for _, s := range sec.KeyStrings() {
		vocDic[s] = sec.Key(s).Value()
	}

	return vocDic, nil
}
