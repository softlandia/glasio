package glasio

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
func (o TWarning) String() string {
	return fmt.Sprintf("line: %d,\tdesc: %s", o.line, o.desc)
}

// ToCsvString - return string with warning
// field TWarning.direct do not write to string
func (o *TWarning) ToCsvString(sep ...string) string {
	var fieldSep string
	switch len(sep) {
	case 0:
		fieldSep = ";"
	case 1:
		fieldSep = sep[0]
	}
	return fmt.Sprintf("%d%s%d%s\"%s\"", o.section, fieldSep, o.line, fieldSep, o.desc)
}

//TLasWarnings - class to store and manipulate warnings
//Count() - return wrnings count
//SaveWarning(fileName string) error
//SaveWarningToWriter(writer *bufio.Writer) int
//SaveWarningToFile(oFile *os.File) int
//ToString() string
//for i, w := range obj {w.ToString()} - перебор всех варнингов
type TLasWarnings []TWarning

//Count - return number of element
func (o TLasWarnings) Count() int {
	return len(o)
}

// ToString - make one string from all elements
// sep[0] - record separator разделитель записей
// sep[1] - field separator разделитель полей
// default separator between field ";" between record "\n"
// on empty container return ""
func (o *TLasWarnings) ToString(sep ...string) string {
	if o.Count() == 0 {
		return ""
	}
	var (
		result   string
		fieldSep string
		recSep   string
	)
	switch len(sep) {
	case 0:
		recSep = "\n"
		fieldSep = ";"
	case 1:
		recSep = sep[0]
		fieldSep = ";"
	case 2:
		recSep = sep[0]
		fieldSep = sep[1]
	default:
		recSep = sep[0]
		fieldSep = sep[1]
	}
	for _, w := range *o {
		result += (w.ToCsvString(fieldSep) + recSep)
	}
	return result
}

//SaveWarning - save to file all warning
//file created and closed
func (o *TLasWarnings) SaveWarning(fileName string) error {
	if o.Count() == 0 {
		return nil
	}
	oFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_ = o.SaveWarningToFile(oFile)
	oFile.Close()
	return nil
}

//SaveWarningToWriter - store all warning to writer
//return count lines writed to
func (o *TLasWarnings) SaveWarningToWriter(writer *bufio.Writer) int {
	if o.Count() == 0 {
		return 0
	}
	for _, w := range *o {
		_, err := writer.WriteString(w.String())
		if err != nil {
			log.Fatal("internal __error__ in SaveWarningToWriter")
		}
	}
	return o.Count()
}

//SaveWarningToFile - store all warning to file, file not close. return count warning writed
func (o *TLasWarnings) SaveWarningToFile(oFile *os.File) int {
	if oFile == nil {
		return 0
	}
	if o.Count() == 0 {
		return 0
	}
	for i, w := range *o {
		fmt.Fprintf(oFile, "%d, %s\n", i, w)
	}
	return o.Count()
}
