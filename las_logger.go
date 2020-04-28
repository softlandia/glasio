//(c) softland 2020
//softlandia@gmail.com

package glasio

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Logger - store log info about the one las, fills up info from las.open() and las.check()
type Logger struct {
	las             *Las         // object from message collected
	filename        string       // file to read, used for reporting // имя файла по которому формируется отчёт, используется для оформления сообщений
	readedNumPoints int          // number points readed from file get from las.Open()
	errorOnOpen     error        // status from las.Open()
	msgOpen         TLasWarnings // сообщения формируемые в процессе открытия las файла
	msgCheck        tCheckMsg    // информация об особых проверках, получаем из LasChecker
	msgCurve        tCurvRprt    // информация о кривых хранящихся в LAS файле, записывается в "log.info.md"
	missMnemonic    tMMnemonic   // мнемоники найденные в файле и не найденные в словаре
}

// NewLogger - constructor
func NewLogger(las *Las) *Logger {
	logger := new(Logger)
	logger.las = las
	logger.filename = las.FileName
	logger.msgOpen = nil
	logger.msgCheck = make(tCheckMsg, 0, 10)
	logger.msgCurve = make(tCurvRprt, 0, 10)
	logger.missMnemonic = make(tMMnemonic)
	return logger
}

// tCheckMsg - хранит все сообщения о специальных проверках las файла
type tCheckMsg []string

func (m tCheckMsg) String() string {
	var sb strings.Builder
	for _, msg := range m {
		sb.WriteString(msg)
	}
	return sb.String()
}

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

func (cr tCurvRprt) String(filename string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("##logs in file: '%s'##\n", filename))
	for _, s := range cr {
		sb.WriteString(s)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

func (cr *tCurvRprt) save(f *os.File, filename string) {
	fmt.Fprintf(f, "##logs in file: '%s'##\n", filename)
	for _, s := range *cr {
		f.WriteString(s)
	}
	f.WriteString("\n")
}

type tMMnemonic map[string]string

func (mm tMMnemonic) String() string {
	keys := make([]string, 0, len(mm))
	for k := range mm {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(mm[k] + "\n")
	}
	return sb.String()
}

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
