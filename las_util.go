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

func cmpLas(correct, las *Las) (res bool) {
	res = (correct.Strt == las.Strt)
	res = res && (correct.Stop == las.Stop)
	res = res && (correct.Step == las.Step)
	res = res && (correct.Null == las.Null)
	res = res && (correct.nPoints == las.nPoints)
	res = res && correct.Logs.Cmp(las.Logs)
	res = res && (len(correct.Logs) == len(las.Logs))
	return res
}

// подразумевается считывание из совершенно корректного файла
// ошибок при заполнении нет, ничего не проверяем
func makeLasFromFile(fn string) *Las {
	las := NewLas()
	las.Open(fn)
	return las
}

func makeSampleLas(
	cp cpd.IDCodePage,
	null float64,
	strt float64,
	stop float64,
	step float64,
	well string) (las *Las) {
	if cp == cpd.CP1251 {
		las = NewLas()
	} else {
		las = NewLas(cp)
	}
	las.Null = null
	las.Strt = strt
	las.Stop = stop
	las.Step = step
	las.Well = well

	curve := NewLasCurve("DEPT.m :")
	curve.Init(len(las.Logs), "DEPT", "DEPT", las.GetExpectedPointsCount())
	las.Logs["DEPT"] = curve
	curve = NewLasCurve("BK.ohmm :laterolog")
	curve.Init(len(las.Logs), "BK", "LL3", las.GetExpectedPointsCount())
	las.Logs["BK"] = curve
	las.setActuallyNumberPoints(5)
	return las
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
