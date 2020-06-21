//(c) softland 2020
//softlandia@gmail.com
package glasio

import (
	"fmt"
	fp "path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMnemonic(t *testing.T) {
	Mnemonic, err := LoadStdMnemonicDic(fp.Join("data/mnemonic.ini"))
	assert.Nil(t, err, fmt.Sprintf("load std mnemonic error: %v\n check out 'data\\mnemonic.ini'", err))

	VocDic, err := LoadStdVocabularyDictionary(fp.Join("data/dic.ini"))
	assert.Nil(t, err, fmt.Sprintf("load std vocabulary dictionary error: %v\n check out 'data\\dic.ini'", err))

	las := NewLas()
	mnemonic := las.GetMnemonic("1")
	assert.Equal(t, mnemonic, "", fmt.Sprintf("<GetMnemonic> return '%s', expected ''\n", mnemonic))

	las.LogDic = &Mnemonic
	las.VocDic = &VocDic

	mnemonic = las.GetMnemonic("GR")
	assert.Equal(t, mnemonic, "GR", fmt.Sprintf("<GetMnemonic> return '%s', expected 'GR'\n", mnemonic))

	mnemonic = las.GetMnemonic("ГК")
	assert.Equal(t, mnemonic, "GR", fmt.Sprintf("<GetMnemonic> return '%s', expected 'GR'\n", mnemonic))
}
