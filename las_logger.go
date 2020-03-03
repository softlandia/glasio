package glasio

import (
	"fmt"
	"os"
	"sort"
)

// LasLog - store log info about the one las, fills up info from las.open() and las.check()
type LasLog struct {
	las             *Las         // object from message collected
	filename        string       // file to read, used for reporting // имя файла по которому формируется отчёт, используется для оформления сообщений
	readedNumPoints int          // number points readed from file get from las.Open()
	errorOnOpen     error        // status from las.Open()
	msgOpen         TLasWarnings // сообщения формируемые в процессе открытия las файла
	msgCheck        tCheckMsg    // информация об особых случаях, получаем из LasChecker
	msgCurve        tCurvRprt    // информация о кривых хранящихся в LAS файле, записывается в "log.info.md"
	missMnemonic    tMMnemonic   // мнемоники найденные в файле и не найденные в словаре
}

// NewLasLog - constructor
func NewLasLog(las *Las) LasLog {
	var lasLog LasLog
	lasLog.las = las
	lasLog.filename = las.FileName
	lasLog.msgOpen = nil
	lasLog.msgCheck = make(tCheckMsg, 0, 10)
	lasLog.msgCurve = make(tCurvRprt, 0, 10)
	lasLog.missMnemonic = make(tMMnemonic, 0)
	return lasLog
}

type tCheckMsg []string

func (m *tCheckMsg) save(f *os.File) {
	for _, msg := range *m {
		f.WriteString(msg)
	}
}

func (m *tCheckMsg) msgFileIsWraped(fn string) string {
	return fmt.Sprintf("file '%s' ignore, WRAP=YES\n", fn)
}

func (m *tCheckMsg) msgFileNoData(fn string) string {
	return fmt.Sprintf("*error* file '%s', no data read ,*ignore*\n", fn)
}

func (m *tCheckMsg) msgFileOpenWarning(fn string, err error) string {
	return fmt.Sprintf("**warning** file '%s' : %v **passed**\n", fn, err)
}

type tCurvRprt []string

func (ir *tCurvRprt) save(f *os.File, filename string) {
	fmt.Fprintf(f, "##logs in file: '%s'##\n", filename)
	for _, s := range *ir {
		f.WriteString(s)
	}
	f.WriteString("\n")
}

type tMMnemonic map[string]string

func (mm *tMMnemonic) save(f *os.File) {
	keys := make([]string, 0, len(*mm))
	for k := range *mm {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f.WriteString((*mm)[k] + "\n")
	}
}
