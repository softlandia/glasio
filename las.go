package glasio

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/softlandia/cpd"

	"github.com/softlandia/xlib"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

///format strings represent structure of LAS file
const (
	_LasFirstLine      = "~VERSION INFORMATION\n"
	_LasVersion        = "VERS.                          %3.1f :glas (c) softlandia@gmail.com\n"
	_LasCodePage       = "CPAGE.                         1251: code page \n"
	_LasWrap           = "WRAP.                          NO  : ONE LINE PER DEPTH STEP\n"
	_LasWellInfoSec    = "~WELL INFORMATION\n"
	_LasMnemonicFormat = "#MNEM.UNIT DATA                                  :DESCRIPTION\n"
	_LasStrt           = " STRT.M %8.3f                                    :START DEPTH\n"
	_LasStop           = " STOP.M %8.3f                                    :STOP  DEPTH\n"
	_LasStep           = " STEP.M %8.3f                                    :STEP\n"
	_LasNull           = " NULL.  %9.3f                                   :NULL VALUE\n"
	_LasRkb            = " RKB.M %8.3f                                     :KB or GL\n"
	_LasXcoord         = " XWELL.M %8.3f                                   :Well head X coordinate\n"
	_LasYcoord         = " YWELL.M %8.3f                                   :Well head Y coordinate\n"
	_LasOilComp        = " COMP.  %-43.43s:OIL COMPANY\n"
	_LasWell           = " WELL.   %-43.43s:WELL\n"
	_LasField          = " FLD .  %-43.43s:FIELD\n"
	_LasLoc            = " LOC .  %-43.43s:LOCATION\n"
	_LasCountry        = " CTRY.  %-43.43s:COUNTRY\n"
	_LasServiceComp    = " SRVC.  %-43.43s:SERVICE COMPANY\n"
	_LasDate           = " DATE.  %-43.43s:DATE\n"
	_LasAPI            = " API .  %-43.43s:API NUMBER\n"
	_LasUwi            = " UWI .  %-43.43s:UNIVERSAL WELL INDEX\n"
	_LasCurvSec        = "~Curve Information Section\n"
	_LasCurvFormat     = "#MNEM.UNIT                 :DESCRIPTION\n"
	_LasCurvDept       = " DEPT.M                    :\n"
	_LasCurvLine       = " %s.%s                     :\n"
	_LasCurvLine2      = " %s                        :\n"
	_LasDataSec        = "~ASCII Log Data\n"
	_LasDataLine       = ""

	//secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - dAta
	lasSecIgnore   = 0
	lasSecVertion  = 1
	lasSecWellInfo = 2
	lasSecCurInfo  = 3
	lasSecData     = 4
)

// Las - class to store las file
// input code page autodetect
// at read file always code page converted to UTF
// at save file code page converted to specifyed in Las.toCodePage
//TODO add pointer to cfg
//TODO warnings - need method to flush slice on file, and clear
//TODO expPoints надо превратить в метод
//TODO имена скважин с пробелами читаются неверно
type Las struct {
	FileName        string              //file name from load
	File            *os.File            //the file from which we are reading
	Reader          io.Reader           //reader created from File, provides decode from codepage to UTF-8
	scanner         *bufio.Scanner      //scanner
	Ver             float64             //version 1.0, 1.2 or 2.0
	Wrap            string              //YES || NO
	Strt            float64             //start depth
	Stop            float64             //stop depth
	Step            float64             //depth step
	Null            float64             //value interpreted as empty
	Well            string              //well name
	Rkb             float64             //altitude KB
	Logs            map[string]LasCurve //store all logs
	LogDic          *map[string]string  //external dictionary of standart log name - mnemonics
	VocDic          *map[string]string  //external vocabulary dictionary of log mnemonic
	Warnings        TLasWarnings        //slice of warnings occure on read or write
	expPoints       int                 //expected count (.)
	nPoints         int                 //actually count (.)
	iCodepage       cpd.IDCodePage      //codepage input file. autodetect
	oCodepage       cpd.IDCodePage      //codepage to save file, default xlib.CpWindows1251. to special value, specify at make: NewLas(cp...)
	iDuplicate      int                 //индекс повторящейся мнемоники, увеличивается на 1 при нахождении дубля, начально 0
	currentLine     int                 //index of current line in readed file
	maxWarningCount int                 //default maximum warning count
	stdNull         float64             //default null value
	CodePage        string              //TODO to delete //пока не читается
}

//GetStepFromData - return step from data section
//read 2 line from section ~A and determine step
//close file
//return o.Null if error occure
//TODO просто сделать функцией
func (o *Las) GetStepFromData(fileName string) float64 {
	iFile, err := os.Open(fileName)
	if err != nil {
		return o.Null
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileStop(fileName, "~A")
	if (err != nil) || (iScanner == nil) {
		return o.Null
	}

	s := ""
	j := 0
	dept1 := 0.0
	dept2 := 0.0
	for i := 0; iScanner.Scan(); i++ {
		s = strings.TrimSpace(iScanner.Text())
		if (len(s) == 0) || (s[0] == '#') {
			continue
		}
		k := strings.IndexRune(s, ' ')
		if k < 0 { //data line must have minimum 2 column separated ' ' space
			return o.Null
		}
		dept1, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			return o.Null
		}
		j++
		if j == 2 {
			return math.Round((dept1-dept2)*10) / 10
		}
		dept2 = dept1
	}
	//если мы попали сюда, то всё грусно, в файле после ~A не нашлось двух строчек с данными... или пустые строчки или комменты
	return o.Null
}

//SetNull - change parameter NULL in WELL INFO section and in all logs
func (o *Las) SetNull(aNull float64) error {
	for _, l := range o.Logs { //loop by logs
		for i := range l.log { //loop by dept step
			if l.log[i] == o.Null {
				l.log[i] = aNull
			}
		}
	}
	o.Null = aNull
	return nil
}

/*
//вызывать после Scanner.Text()
func (o *Las) convertStrFromIn(s string) string {
	switch o.iCodepage {
	case cpd.CP866:
		s, _, _ = transform.String(charmap.CodePage866.NewDecoder(), s)
	case cpd.CP1251:
		s, _, _ = transform.String(charmap.Windows1251.NewDecoder(), s)
	}
	return s
}*/

//TODO replace to function xStrUtil.ConvertStrCodePage
func (o *Las) convertStrToOut(s string) string {
	switch o.oCodepage {
	case cpd.CP866:
		s, _, _ = transform.String(charmap.CodePage866.NewEncoder(), s)
	case cpd.CP1251:
		s, _, _ = transform.String(charmap.Windows1251.NewEncoder(), s)
	}
	return s
}

//logByIndex - return log from map by Index
func (o *Las) logByIndex(i int) (*LasCurve, error) {
	for _, v := range o.Logs {
		if v.Index == i {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("log with index: %v not present", i)
}

//NewLas - make new object Las class
//autodetect code page at load file
//code page to save by default is xlib.CpWindows1251
func NewLas(outputCP ...cpd.IDCodePage) *Las {
	las := new(Las)
	las.Ver = 2.0
	las.Wrap = "NO"
	las.Logs = make(map[string]LasCurve)
	las.maxWarningCount = 20 //TODO read from Cfg
	las.stdNull = -999.25    //TODO read from Cfg
	if len(outputCP) > 0 {
		las.oCodepage = outputCP[0]
	} else {
		las.oCodepage = cpd.CP1251
	}
	//mnemonic dictionary
	las.LogDic = nil
	//external log dictionary
	las.VocDic = nil
	//счётчик повторяющихся мнемоник, увеличивается каждый раз на 1, используется при переименовании мнемоники
	las.iDuplicate = 0
	return las
}

//analize first char after ~
//~V - section vertion
//~W - well info section
//~C - curve info section
//~A - data section
func (o *Las) selectSection(r rune) int {
	switch r {
	case 86: //V
		return lasSecVertion //version section
	case 118: //v
		return lasSecVertion //version section
	case 87: //W
		return lasSecWellInfo //well info section
	case 119: //w
		return lasSecWellInfo //well info section
	case 67: //C
		return lasSecCurInfo //curve section
	case 99: //c
		return lasSecCurInfo //curve section
	case 65: //A
		return lasSecData //data section
	case 97: //a
		return lasSecData //data section
	default:
		return lasSecIgnore
	}
}

//make test of loaded well info section
//return error <> nil in one case, if getStepFromData return error
func (o *Las) testWellInfo() error {
	if o.Step == 0.0 {
		o.Step = o.GetStepFromData(o.FileName) // return o.Null if cannot calculate step from data
		if o.Step == o.Null {
			return errors.New("invalid STEP parameter, equal 0. and invalid step in data")
		}
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid STEP parameter, equal 0. replace to %4.3f", o.Step)})
	}
	if o.Null == 0.0 {
		o.Null = o.stdNull
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid NULL parameter, equal 0. replace to %4.3f", o.Null)})
	}
	if math.Abs(o.Stop-o.Strt) < 0.1 {
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid STRT: %4.3f or STOP: %4.3f, will be replace to actually", o.Strt, o.Stop)})
	}
	return nil
}

// IsWraped - return true if WRAP == YES
func (o *Las) IsWraped() bool {
	return (strings.Index(strings.ToUpper(o.Wrap), "Y") >= 0)
}

// SaveWarning - save to file all warning
func (o *Las) SaveWarning(fileName string) error {
	if o.Warnings.Count() == 0 {
		return nil
	}
	oFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	o.SaveWarningToFile(oFile)
	oFile.Close()
	return nil
}

// SaveWarningToWriter - store all warning to writer, return count lines writed to
func (o *Las) SaveWarningToWriter(writer *bufio.Writer) int {
	n := o.Warnings.Count()
	if n == 0 {
		return 0
	}
	for _, w := range o.Warnings {
		writer.WriteString(w.String())
		writer.WriteString("\n")
	}
	return n
}

// SaveWarningToFile - store all warning to file, file not close. return count warning writed
func (o *Las) SaveWarningToFile(oFile *os.File) int {
	if oFile == nil {
		return 0
	}
	if o.Warnings.Count() == 0 {
		return 0
	}
	oFile.WriteString("**file: " + o.FileName + "**\n")
	n := o.Warnings.SaveWarningToFile(oFile)
	oFile.WriteString("\n")
	return n
}

func (o *Las) addWarning(w TWarning) {
	if o.Warnings.Count() < o.maxWarningCount {
		o.Warnings = append(o.Warnings, w)
		if o.Warnings.Count() == o.maxWarningCount {
			o.Warnings = append(o.Warnings, TWarning{0, 0, 0, "*maximum count* of warning reached, change parameter 'maxWarningCount' in 'glas.ini'"})
		}
	}
}

// GetMnemonic - return Mnemonic from dictionary by Log Name, if Mnemonic not found return empty string ""
func (o *Las) GetMnemonic(logName string) string {
	if (o.LogDic == nil) || (o.VocDic == nil) {
		return "-"
	}
	_, ok := (*o.LogDic)[logName]
	if ok { //GOOD - название каротажа равно мнемонике
		return logName
	}
	v, ok := (*o.VocDic)[logName]
	if ok { //POOR - название загружаемого каротажа найдено в словаре подстановок, мнемоника найдена
		return v
	}
	return ""
}

// Open - load las file
func (o *Las) Open(fileName string) (int, error) {
	//TODO при создании объекта las есть возможность указать кодировку записи, нужна возможность указать явно кодировку чтения
	var err error
	o.File, err = os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer o.File.Close()
	o.FileName = fileName
	//store Reader, this reader decode to UTF-8
	o.Reader, err = cpd.NewReader(o.File)
	if err != nil {
		return 0, err
	}
	o.scanner = bufio.NewScanner(o.Reader)
	/*o.iCodepage, err = cpd.FileCodepageDetect(fileName)*/

	//load header from stored Reader
	o.currentLine = 0
	err = o.LoadHeader()
	if err != nil {
		return 0, err
	}

	if o.IsWraped() {
		o.addWarning(TWarning{directOnRead, lasSecData, -1, "WRAP = YES, file ignored"})
		return 0, nil
	}
	if len(o.Logs) <= 0 {
		o.addWarning(TWarning{directOnRead, lasSecData, -1, "section ~Curve not exist, file ignored"})
		return 0, nil
	}
	return o.ReadDataSec(fileName)
}

//LoadHeader - read las file and load all section before dAta ~A
/*  secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - A data
1. читаем строку
2. если коммент или пустая в игнор
3. если начало секции, определяем какой
4. если началась секция данных заканчиваем
5. читаем одну строку (это один параметер из известной нам секции) */
func (o *Las) LoadHeader() error {
	s := ""
	var err error
	secNum := 0
	for i := 0; o.scanner.Scan(); i++ {
		s = strings.TrimSpace(o.scanner.Text())
		o.currentLine++
		if isIgnoredLine(s) {
			continue
		}
		if s[0] == '~' { //start new section
			secNum = o.selectSection(rune(s[1]))
			if secNum == lasSecCurInfo { //enter to Curve section.
				//проверка корректности данных секции WELL INFO перез загрузкой кривых и данных
				//TODO проверку нужно перенести в функцию Open() здесь вообще не место
				err = o.testWellInfo() //STEP != 0, NULL != 0, STRT & STOP
				if err != nil {
					return err // двойная ошибка, плох параметр STEP и не удалось вычислить STEP по данным, с данными проблема...
				}
			}
			if secNum == lasSecData {
				break // dAta section read after //exit from for
			}
		} else {
			err = o.ReadParameter(s, secNum) //if not comment, not empty and not new section => parameter, read it
			if err != nil {
				o.addWarning(TWarning{directOnRead, secNum, -1, fmt.Sprintf("while process parameter: '%s' occure error: %v", s, err)})
			}
		}
	}
	return nil
}

//ReadParameter - read one parameter
func (o *Las) ReadParameter(s string, secNum int) error {
	switch secNum {
	case lasSecVertion:
		return o.readVersionParam(s)
	case lasSecWellInfo:
		return o.ReadWellParam(s)
	case lasSecCurInfo:
		return o.readCurveParam(s)
	}
	return nil
}

func (o *Las) readVersionParam(s string) error {
	var err error
	p := NewLasParamFromString(s)
	switch p.Name {
	case "VERS":
		o.Ver, err = strconv.ParseFloat(p.Val, 64)
	case "WRAP":
		o.Wrap = p.Val
	}
	return err
}

//ReadWellParam - read parameter from WELL section
func (o *Las) ReadWellParam(s string) error {
	var err error
	p := NewLasParamFromString(s)
	switch p.Name {
	case "STRT":
		o.Strt, err = strconv.ParseFloat(p.Val, 64)
	case "STOP":
		o.Stop, err = strconv.ParseFloat(p.Val, 64)
	case "STEP":
		o.Step, err = strconv.ParseFloat(p.Val, 64)
	case "NULL":
		o.Null, err = strconv.ParseFloat(p.Val, 64)
	case "WELL":
		if o.Ver < 2.0 {
			o.Well = p.Desc
		} else {
			o.Well = p.Val
		}
	}
	if err != nil {
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("detected param: %v, unit:%v, value: %v\n", p.Name, p.Unit, p.Val)})
	}
	return err
}

//ChangeDuplicateLogName - return non duplicated name of log
//if input name unique, return input name
//if input name not unique, return input name + index duplicate
//index duplicate - Las field, increase
func (o *Las) ChangeDuplicateLogName(name string) string {
	s := ""
	if _, ok := o.Logs[name]; ok {
		o.iDuplicate++
		s = fmt.Sprintf("%v", o.iDuplicate)
		name += s
	}
	return name
}

//Разбор одной строки с мнемоникой каротажа
//Разбираем в переменную l а потом сохраняем в map
//Каждый каротаж характеризуется тремя именами
//IName    - имя каротажа в исходном файле, может повторятся
//Name     - ключ в map хранилище, повторятся не может. если в исходном есть повторение, то Name строится добавлением к IName индекса
//Mnemonic - мнемоника, берётся из словаря. если в словаре не найдено, то оставляем iName
func (o *Las) readCurveParam(s string) error {
	l := NewLasCurveFromString(s)
	l.Init(len(o.Logs), o.GetMnemonic(l.Name), o.ChangeDuplicateLogName(l.Name), o.GetExpectedPointsCount())
	o.Logs[l.Name] = l //добавление в карту кривой каротажа с колонкой глубин
	return nil
}

//GetExpectedPointsCount - оценка количества точек по параметрам STEP, STRT, STOP
func (o *Las) GetExpectedPointsCount() int {
	var m int
	if math.Abs(o.Stop) > math.Abs(o.Strt) {
		m = int((o.Stop-o.Strt)/o.Step) + 2
	} else {
		m = int((o.Strt-o.Stop)/o.Step) + 2
	}
	if m < 0 {
		m = -m
	}
	return m
}

//expandDept - if actually data points exceeds
func (o *Las) expandDept(d *LasCurve) {
	//actual number of points more then expected
	o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "actual number of data lines more than expected, check: STRT, STOP, STEP"})
	o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "expand number of points"})
	//ожидаем удвоения данных

	o.expPoints *= 2
	//need expand all logs
	//fmt.Printf("old dept len: %d, cap: %d\n", len(d.dept), cap(d.dept))

	newDept := make([]float64, o.expPoints, o.expPoints)
	copy(newDept, d.dept)
	d.dept = newDept

	newLog := make([]float64, o.expPoints, o.expPoints)
	copy(newLog, d.dept)
	d.log = newLog
	o.Logs[d.Name] = *d

	//fmt.Printf("new dept len: %d, cap: %d\n", len(d.dept), cap(d.dept))
	//loop over other logs
	n := len(o.Logs)
	var l *LasCurve
	for j := 1; j < n; j++ {
		l, _ = o.logByIndex(j)
		newDept := make([]float64, o.expPoints, o.expPoints)
		copy(newDept, l.dept)
		l.dept = newDept

		newLog := make([]float64, o.expPoints, o.expPoints)
		copy(newLog, l.log)
		l.log = newLog
		o.Logs[l.Name] = *l
	}
}

// ReadDataSec - read section of data
// TODO file open and not close
func (o *Las) ReadDataSec(fileName string) (int, error) {
	var (
		v    float64
		err  error
		d    *LasCurve
		l    *LasCurve
		dept float64
		i    int
	)

	/*
		   //обнаруживаем начало секции данных, позиционируемся на строку перед данными
		   	pos, iScanner, err := xlib.SeekFileStop(fileName, "~A")

		   	//now current position in file at line "~A..."
		   	o.currentLine = pos

		switch pos {
		case 0:
			return 0, err
		case -1:
			return 0, fmt.Errorf("<ReadDataSec> data section '~A' not found")
		}
	*/
	//iScanner := o.scanner

	//исходя из параметров STRT, STOP и STEP определяем ожидаемое количество строк данных
	o.expPoints = o.GetExpectedPointsCount()
	//o.currentLine++
	n := len(o.Logs)       //количество каротажей, столько колонок данных ожидаем
	d, _ = o.logByIndex(0) //dept log
	s := ""
	for i = 0; o.scanner.Scan(); i++ {
		o.currentLine++
		if i == o.expPoints {
			o.expandDept(d)
		}
		s = strings.TrimSpace(o.scanner.Text())
		if isIgnoredLine(s) {
			i--
			continue
		}
		//first column is DEPT
		k := strings.IndexRune(s, ' ')
		if k < 0 { //line must have n+1 column and n separated spaces block (+1 becouse first column DEPT)
			o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("line: %d is empty, ignore", o.currentLine)})
			i--
			continue
		}
		dept, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("first column '%s' not numeric, ignore", s[:k])})
			i--
			continue
		}

		d.dept[i] = dept
		if i > 1 {
			if math.Pow(((dept-d.dept[i-1])-o.Step), 2) > 0.1 {
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("actual step %5.2f ≠ global STEP %5.2f", (dept - d.dept[i-1]), o.Step)})
			}
		}
		if i > 2 {
			if math.Pow(((dept-d.dept[i-1])-(d.dept[i-1]-d.dept[i-2])), 2) > 0.1 {
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("step %5.2f ≠ previously step %5.2f", (dept - d.dept[i-1]), (d.dept[i-1] - d.dept[i-2]))})
				dept = d.dept[i-1] + o.Step
			}
		}

		s = strings.TrimSpace(s[k+1:]) //cut first column
		//цикл по каротажам
		for j := 1; j < (n - 1); j++ {
			iSpace := strings.IndexRune(s, ' ')
			switch iSpace {
			case -1: //не все колонки прочитаны, а пробелов уже нет... пробуем игнорировать сроку заполняя оставшиеся каротажи NULLами
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "not all column readed, set log value to NULL"})
			case 0:
				v = o.Null
			case 1:
				v, err = strconv.ParseFloat(s[:1], 64)
			default:
				v, err = strconv.ParseFloat(s[:iSpace], 64) //strconv.ParseFloat(s[:iSpace-1], 64)
			}
			if err != nil {
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("can't convert string: '%s' to number, set to NULL", s[:iSpace-1])})
				v = o.Null
			}
			l, err = o.logByIndex(j)
			if err != nil {
				o.nPoints = i
				return i, errors.New("internal ERROR, func (o *Las) readDataSec()::o.logByIndex(j) return error")
			}
			l.dept[i] = dept
			l.log[i] = v
			s = strings.TrimSpace(s[iSpace+1:])
		}
		//остаток - последняя колонка
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "not all column readed, set log value to NULL"})
			v = o.Null
		}
		l, err = o.logByIndex(n - 1)
		if err != nil {
			o.nPoints = i
			return i, errors.New("internal ERROR, func (o *Las) readDataSec()::o.logByIndex(j) return error on last column")
		}
		l.dept[i] = dept
		l.log[i] = v
	}
	//i - actually readed lines and add (.) to data array
	//crop logs to actually len
	o.setActuallyNumberPoints(i)
	return i, nil
}

// NumPoints - return actually number of points in data
func (o *Las) NumPoints() int {
	return o.nPoints
}

//Dept - return slice of DEPT curve (first column)
func (o *Las) Dept() []float64 {
	d, err := o.logByIndex(0)
	if err != nil {
		return nil
	}
	return d.dept
}

func (o *Las) setActuallyNumberPoints(numPoints int) error {
	if numPoints <= 0 {
		o.nPoints = 0
		return errors.New("internal ERROR, func (o *Las) setActuallyNumberPoints(), actually number of points <= 0")
	}
	if numPoints > len(o.Dept()) {
		o.nPoints = 0
		return errors.New("internal ERROR, func (o *Las) setActuallyNumberPoints(), actually number of points > then exist data")
	}
	for _, l := range o.Logs {
		l.SetLen(numPoints)
	}
	o.nPoints = numPoints
	return nil
}

//Save - save to file
//rewrite if file exist
//if useMnemonic == true then on save using std mnemonic on ~Curve section
//TODO las have field filename of readed las file, after save filename must update or not? warning occure on write for what file?
func (o *Las) Save(fileName string, useMnemonic ...bool) error {
	n := len(o.Logs) //log count
	if n <= 0 {
		return errors.New("logs not exist")
	}

	var f *os.File
	var err error
	if !xlib.FileExists(fileName) {
		err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
		if err != nil {
			return errors.New("path: '" + filepath.Dir(fileName) + "' can't create >>" + err.Error())
		}
	}
	f, err = os.Create(fileName) //Open file to WRITE
	if err != nil {
		return errors.New("file: '" + fileName + "' can't open to write >>" + err.Error())
	}
	defer f.Close()

	fmt.Fprint(f, _LasFirstLine)
	fmt.Fprintf(f, _LasVersion, o.Ver)
	fmt.Fprint(f, _LasWrap)
	fmt.Fprint(f, _LasCodePage)
	fmt.Fprint(f, _LasWellInfoSec)
	fmt.Fprintf(f, _LasStrt, o.Strt)
	fmt.Fprintf(f, _LasStop, o.Stop)
	fmt.Fprintf(f, _LasStep, o.Step)
	fmt.Fprintf(f, _LasNull, o.Null)
	fmt.Fprintf(f, _LasWell, o.convertStrToOut(o.Well))
	fmt.Fprint(f, _LasCurvSec)
	fmt.Fprint(f, _LasCurvDept)

	s := "# DEPT  |" //готовим строчку с названиями каротажей глубина всегда присутствует
	var l *LasCurve
	for i := 1; i < n; i++ { //Пишем названия каротажей
		l, _ := o.logByIndex(i)
		if len(useMnemonic) > 0 {
			if len(l.Mnemonic) > 0 {
				l.Name = l.Mnemonic
			}
		}
		fmt.Fprintf(f, _LasCurvLine, o.convertStrToOut(l.Name), o.convertStrToOut(l.Unit)) //запись мнемоник в секции ~Curve
		s += " " + fmt.Sprintf("%-8s|", l.Name)                                            //Собираем строчку с названиями каротажей
	}

	fmt.Fprintf(f, _LasDataSec)
	//write data
	s += "\n"
	fmt.Fprintf(f, o.convertStrToOut(s))
	dept, _ := o.logByIndex(0)
	for i := 0; i < o.nPoints; i++ { //loop by dept (.)
		fmt.Fprintf(f, "%-9.3f ", dept.dept[i])
		for j := 1; j < n; j++ { //loop by logs
			l, err = o.logByIndex(j)
			if err != nil {
				o.addWarning(TWarning{directOnWrite, lasSecData, i, "logByIndex() return error, log not found, panic"})
				return errors.New("logByIndex() return error, log not found, panic")
			}
			fmt.Fprintf(f, "%-9.3f ", l.log[i])
		}
		fmt.Fprintln(f)
	}
	return nil
}
