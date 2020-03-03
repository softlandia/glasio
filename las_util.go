package glasio

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/softlandia/cpd"
)

func isIgnoredLine(s string) bool {
	if (len(s) == 0) || (s[0] == '#') {
		return true
	}
	return false
}

// LoadLasHeader - utility function, if need read only header without data
func LoadLasHeader(fileName string) (*Las, error) {
	las := NewLas()
	iFile, err := os.Open(fileName)
	if err != nil {
		return nil, errors.New("could not open file: '" + fileName + "'")
	}
	defer iFile.Close()
	las.File = iFile
	las.FileName = fileName
	las.Reader, err = cpd.NewReader(las.File)
	las.scanner = bufio.NewScanner(las.Reader)
	if err != nil {
		return nil, err
	}
	las.LoadHeader()
	return las, nil
}

// считывает файл и собирает все сообщения в один объект
func lasOpenCheck(filename string) LasLog {

	las := NewLas() // TODO make special constructor to initialize with global Mnemonic and Dic
	//las.LogDic = &Mnemonic // global var
	//las.VocDic = &Dic      // global var

	LasLog := NewLasLog(las)

	LasLog.readedNumPoints, LasLog.errorOnOpen = las.Open(filename)
	LasLog.msgOpen = las.Warnings

	if las.IsWraped() {
		LasLog.msgCheck = append(LasLog.msgCheck, LasLog.msgCheck.msgFileIsWraped(filename))
		//return statLasCheck_WRAP
	}
	if las.NumPoints() == 0 {
		LasLog.msgCheck = append(LasLog.msgCheck, LasLog.msgCheck.msgFileNoData(filename))
		//return statLasCheck_DATA
	}
	if LasLog.errorOnOpen != nil {
		LasLog.msgCheck = append(LasLog.msgCheck, LasLog.msgCheck.msgFileOpenWarning(filename, LasLog.errorOnOpen))
	}

	for k, v := range las.Logs {
		if len(v.Mnemonic) == 0 { //v.Mnemonic содержит автоопределённую стандартную мнемонику, если она пустая, значит пропущена, помечаем **
			LasLog.msgCurve = append(LasLog.msgCurve, fmt.Sprintf("*input log: %s \t internal: %s \t mnemonic:%s*\n", v.IName, k, v.Mnemonic))
			LasLog.missMnemonic[v.IName] = v.IName
		} else {
			LasLog.msgCurve = append(LasLog.msgCurve, fmt.Sprintf("input log: %s \t internal: %s \t mnemonic: %s\n", v.IName, k, v.Mnemonic))
		}
	}

	las = nil
	return LasLog
}
