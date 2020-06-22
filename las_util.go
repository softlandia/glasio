package glasio

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/softlandia/cpd"
)

// isIgnoredLine - check string s to empty or LAS format comment
func isIgnoredLine(s string) bool {
	if (len(s) == 0) || (s[0] == '#') {
		return true
	}
	return false
}

// два las равны если равны их основные параметры: STRT, STOP, STEP, NULL, количество точек в данных,
// а также количество кривых и совпадают имена кривых
func cmpLas(correct, las *Las) (res bool) {
	res = (correct.Strt == las.Strt)
	res = res && (correct.Stop == las.Stop)
	res = res && (correct.Step == las.Step)
	res = res && (correct.Null == las.Null)
	res = res && (correct.NumPoints() == las.NumPoints())
	res = res && correct.Logs.Cmp(las.Logs)
	return res
}

// подразумевается считывание из совершенно корректного файла
// ошибок при заполнении нет, ничего не проверяем
func makeLasFromFile(fn string) *Las {
	las := NewLas()
	las.Open(fn)
	return las
}

// create object *Las
// NULL, STRT, STOP, STEP, WELL from input
// create 2 curves: DEPT and BK with 5 points
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

	curve := NewLasCurve("DEPT.m :", las)
	curve.D = append(curve.D, 1.0, 1.1, 1.2, 1.3, 1.4)
	las.Logs = append(las.Logs, curve)

	curve = NewLasCurve("BK.ohmm :laterolog", las)
	curve.D = append(curve.D, 1.0, 1.1, 1.2, 1.3, 1.4)
	curve.V = append(curve.V, 0.0, 1.1, 2.2, 3.3, 4.4)
	las.Logs = append(las.Logs, curve)
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
	//las.ePoints = las.ReadRows()
	las.ReadRows()
	las.LoadHeader()
	return las, nil
}

// LasCheck - read and check las file, return object with all warnings
// считывает файл и собирает все сообщения в один объект
// это базовая проверка las файла, прелесть в том что здесь собираются сообщения от прочтения файла
func LasCheck(filename string) *Logger {
	las := NewLas()
	n, err := las.Open(filename)
	lasLog := NewLogger(las)
	lasLog.readedNumPoints = n
	lasLog.errorOnOpen = err
	lasLog.msgOpen = las.Warnings
	if las.IsWraped() {
		lasLog.msgCheck = append(lasLog.msgCheck, lasLog.msgCheck.msgFileIsWraped(filename))
	}
	if las.NumPoints() == 0 {
		lasLog.msgCheck = append(lasLog.msgCheck, lasLog.msgCheck.msgFileNoData(filename))
	}
	if lasLog.errorOnOpen != nil {
		lasLog.msgCheck = append(lasLog.msgCheck, lasLog.msgCheck.msgFileOpenWarning(filename, lasLog.errorOnOpen))
	}
	return lasLog
}

// LasDeepCheck - read and check las file, curve name checked to mnemonic, return object with all warnings
// считывает файл и собирает все сообщения в один объект
func LasDeepCheck(filename, mnemonicFile, vocdicFile string) (*Logger, error) {
	lasLog := LasCheck(filename)
	if lasLog.errorOnOpen != nil {
		//при выполнении LasCheck произошла ошибка чтения файла, дальнейшаа более глубокая проверка нежелательна
		//ошибки чтения файла связаны с серьёздным нарушением его структуры, углубленная проверка не имеет смысла
		return lasLog, lasLog.errorOnOpen
	}
	//TODO здесь засада, LasCheck сам создаёт и читает las, более того он вообще-то его в себе хранит,
	//     НО в данном случае нам СТОИТ??? или НЕ СТОИТ??? об этом забывать
	//     мы ведь вынуждены всё равно прочитать ещё раз las файл
	las := NewLas()
	Mnemonic, err := LoadStdMnemonicDic(fp.Join(mnemonicFile))
	if err != nil {
		return nil, err
	}
	VocDic, err := LoadStdVocabularyDictionary(fp.Join(vocdicFile))
	if err != nil {
		return nil, err
	}
	las.LogDic = &Mnemonic
	las.VocDic = &VocDic
	las.Open(filename) //читаем второй раз, когда подключены словари, то чтение идёт иначе )))
	for _, curve := range las.Logs {
		if len(curve.Mnemonic) == 0 { //curve.Mnemonic содержит автоопределённую стандартную мнемонику, если она пустая, значит пропущена, помечаем **
			lasLog.msgCurve = append(lasLog.msgCurve, fmt.Sprintf("*input log: %s \t mnemonic:%s*\n", curve.IName, curve.Mnemonic))
			lasLog.missMnemonic[curve.IName] = curve.IName
		} else {
			lasLog.msgCurve = append(lasLog.msgCurve, fmt.Sprintf("input log: %s \t mnemonic: %s\n", curve.IName, curve.Mnemonic))
		}
	}
	las = nil
	return lasLog, nil
}
