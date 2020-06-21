package glasio

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	//defWarningCount = 10
	//warningUNDEF  = 0
	directOnRead  = 1
	directOnWrite = 2
)

//TWarning - class to store warning
type TWarning struct {
	direct  int    // 0 - undefine (warningUNDEF), 1 - on read (directOnRead), 2 - on write (directOnWrite)
	section int    // 0 - undefine (warningUNDEF), lasSecVertion, lasSecWellInfo, lasSecCurInfo, lasSecData
	line    int    // number of line in source file
	desc    string // description of warning
}

//String - return string with warning
func (w TWarning) String() string {
	return fmt.Sprintf("line: %d,\t\"%s\"", w.line+1, w.desc)
}

// ToCsvString - return string with warning
// field TWarning.direct do not write to string
func (w *TWarning) ToCsvString(sep ...string) string {
	var fieldSep string
	switch len(sep) {
	case 0:
		fieldSep = ";"
	case 1:
		fieldSep = sep[0]
	}
	return fmt.Sprintf("%3d%s \"%s\"", w.line+1, fieldSep, w.desc)
}

//TLasWarnings - class to store and manipulate warnings
//Count() - return warnings count
//SaveWarning(fileName string) error
//SaveWarningToWriter(writer *bufio.Writer) int
//SaveWarningToFile(oFile *os.File) int
//ToString() string
//for i, w := range obj {w.ToString()} - перебор всех варнингов
type TLasWarnings []TWarning

//separators for output Warnings to string
var (
	RecordSeparator = "\n"
	FieldSeparator  = ","
)

//Count - return number of element
func (w TLasWarnings) Count() int {
	return len(w)
}

// ToString - make one string from all elements
// sep[0] - record separator разделитель записей
// sep[1] - field separator разделитель полей
// default separator between field "," between record "\n"
// on empty container return ""
func (w *TLasWarnings) ToString(sep ...string) string {
	if w.Count() == 0 {
		return ""
	}
	var (
		fieldSep string
		recSep   string
	)
	switch len(sep) {
	case 0:
		recSep = RecordSeparator
		fieldSep = FieldSeparator
	case 1:
		recSep = sep[0]
		fieldSep = FieldSeparator
	case 2:
		recSep = sep[0]
		fieldSep = sep[1]
	default:
		recSep = sep[0]
		fieldSep = sep[1]
	}
	var sb strings.Builder
	for i, wrn := range *w {
		sb.WriteString(fmt.Sprintf("%2d%s %s%s", i, fieldSep, wrn.ToCsvString(fieldSep), recSep))
	}
	return sb.String()
}

//SaveWarning - save to file all warning
//file created and closed
func (w *TLasWarnings) SaveWarning(fileName string) error {
	if w.Count() == 0 {
		return nil
	}
	oFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_ = w.SaveWarningToFile(oFile)
	oFile.Close()
	return nil
}

//SaveWarningToWriter - store all warning to writer
//return count lines writed to
func (w *TLasWarnings) SaveWarningToWriter(writer *bufio.Writer) int {
	if w.Count() == 0 {
		return 0
	}
	for _, w := range *w {
		_, err := writer.WriteString(w.String())
		if err != nil {
			log.Fatal("internal __error__ in SaveWarningToWriter")
		}
	}
	return w.Count()
}

//SaveWarningToFile - store all warning to file, file not close. return count warning writed
func (w *TLasWarnings) SaveWarningToFile(oFile *os.File) int {
	if oFile == nil {
		return 0
	}
	if w.Count() == 0 {
		return 0
	}
	for i, wrn := range *w {
		fmt.Fprintf(oFile, "%d, %s\n", i, wrn)
	}
	return w.Count()
}
