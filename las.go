// (c) softland 2020
// softlandia@gmail.com
// main file

package glasio

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/softlandia/cpd"

	"github.com/softlandia/xlib"
)

// Las - class to store las file
// input code page autodetect
// at read file always code page converted to UTF
// at save file code page converted to specifyed in Las.oCodepage
//TODO add pointer to cfg
//TODO при создании объекта las есть возможность указать кодировку записи, нужна возможность указать явно кодировку чтения
type Las struct {
	rows            []string           // buffer for read source file, only converted to UTF-8 no any othe change
	FileName        string             // file name from load
	File            *os.File           // the file from which we are reading
	Reader          io.Reader          // reader created from File, provides decode from codepage to UTF-8
	scanner         *bufio.Scanner     // scanner
	Logs            LasCurves          // store all logs
	LogDic          *map[string]string // external dictionary of standart log name - mnemonics
	VocDic          *map[string]string // external vocabulary dictionary of log mnemonic
	Warnings        TLasWarnings       // slice of warnings occure on read or write
	oCodepage       cpd.IDCodePage     // codepage to save, default xlib.CpWindows1251. to special value, specify at make: NewLas(cp...)
	currentLine     int                // index of current line in readed file
	maxWarningCount int                // default maximum warning count
	stdNull         float64            // default null value
	VerSec,
	WelSec,
	CurSec,
	ParSec,
	OthSec HeaderSection
}

var (
	// ExpPoints - первоначальный размер слайсов для хранения данных Logs.D и Logs.V
	// ожидаемое количество точек данных, до чтения мы не можем знать сколько точек будет фактически прочитано
	// данные увеличиваются при необходимости и обрезаются после окончания чтения
	ExpPoints int = 1000
	// StdNull - пустое значение
	StdNull float64 = -999.25
	// MaxWarningCount - слишком много сообщений писать смысла нет
	MaxWarningCount int = 20
)

//method for get values from header containers ////////////////

func (las *Las) parFloat(sec HeaderSection, name string, defValue float64) float64 {
	v := defValue
	if p, ok := sec.params[name]; ok {
		v, _ = strconv.ParseFloat(p.Val, 64)
	}
	return v
}

func (las *Las) parStr(sec HeaderSection, name, defValue string) string {
	v := defValue
	if p, ok := sec.params[name]; ok {
		v = p.Val
	}
	return v
}

// NULL - return null value of las file as float64
// if parameter NULL in las file not exist, then return StdNull (by default -999.25)
func (las *Las) NULL() float64 {
	return las.parFloat(las.WelSec, "NULL", StdNull)
}

// STOP - return depth stop value of las file as float64
// if parameter STOP in las file not exist, then return StdNull (by default -999.25)
func (las *Las) STOP() float64 {
	return las.parFloat(las.WelSec, "STOP", StdNull)
}

// STRT - return depth start value of las file as float64
// if parameter NULL in las file not exist, then return StdNull (by default -999.25)
func (las *Las) STRT() float64 {
	return las.parFloat(las.WelSec, "STRT", StdNull)
}

// STEP - return depth step value of las file as float64
// if parameter not exist, then return StdNull (by default -999.25)
func (las *Las) STEP() float64 {
	return las.parFloat(las.WelSec, "STEP", StdNull)
}

// VERS - return version of las file as float64
// if parameter VERS in las file not exist, then return 2.0
func (las *Las) VERS() float64 {
	return las.parFloat(las.VerSec, "VERS", 2.0)
}

// WRAP - return wrap parameter of las file
// if parameter not exist, then return "NO"
func (las *Las) WRAP() string {
	return las.parStr(las.VerSec, "WRAP", "NO")
}

// WELL - return well name
// if parameter WELL in las file not exist, then return "--"
func (las *Las) WELL() string {
	return las.parStr(las.WelSec, "WELL", "")
}

//NewLas - make new object Las class
//autodetect code page at load file
//code page to save by default is cpd.CP1251
func NewLas(outputCP ...cpd.IDCodePage) *Las {
	las := new(Las)
	las.rows = make([]string, 0, ExpPoints /*las.nRows*/)
	las.Logs = make([]LasCurve, 0)
	las.VerSec = NewVerSection()
	las.WelSec = NewWelSection()
	las.CurSec = NewCurSection()
	las.ParSec = NewParSection()
	las.OthSec = NewOthSection()
	las.maxWarningCount = MaxWarningCount
	las.stdNull = StdNull
	if len(outputCP) > 0 {
		las.oCodepage = outputCP[0]
	} else {
		las.oCodepage = cpd.CP1251
	}
	//mnemonic dictionary
	las.LogDic = nil
	//external log dictionary
	las.VocDic = nil
	return las
}

// IsWraped - return true if WRAP == YES
func (las *Las) IsWraped() bool {
	return strings.Contains(strings.ToUpper(las.WRAP()), "Y")
}

// GetRows - get internal field 'rows'
func (las *Las) GetRows() []string {
	return las.rows
}

// ReadRows - reads to buffer 'rows' and return total count of read lines
func (las *Las) ReadRows() int {
	for i := 0; las.scanner.Scan(); i++ {
		las.rows = append(las.rows, las.scanner.Text())
	}
	return len(las.rows)
}

// Load - load las from reader
// you can make reader from string or othe containers and send as input parameters
func (las *Las) Load(reader io.Reader) (int, error) {
	var err error
	if reader == nil {
		return 0, errors.New("Load received nil reader")
	}
	//create Reader, this reader decode to UTF-8 from reader
	las.Reader, err = cpd.NewReader(reader)
	if err != nil {
		return 0, err //FATAL error - file cannot be decoded to UTF-8
	}
	// prepare file to read
	las.scanner = bufio.NewScanner(las.Reader)
	las.ReadRows()
	m, _ := las.LoadHeader()
	stdChecker := NewStdChecker()
	// check for FATAL errors
	r := stdChecker.check(las)
	las.storeHeaderWarning(r)
	if err = r.fatal(); err != nil {
		return 0, err
	}
	if r.nullWrong() {
		las.SetNull(las.stdNull)
	}
	if r.strtWrong() {
		h := las.GetStrtFromData() // return las.Null if cannot find strt in the data section.
		if h == las.NULL() {
			las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprint("__WRN__ STRT parameter on data is wrong setting to 0")})
			las.setStrt(0)
		}
		las.setStrt(h)
	}
	if r.stepWrong() {
		h := las.GetStepFromData() // return las.Null if cannot calculate step from data
		if h == las.NULL() {
			las.addWarning(TWarning{directOnRead, lasSecWellInfo, las.currentLine, fmt.Sprint("__WRN__ STEP parameter on data is wrong")})
		}
		las.setStep(h)
	}
	return las.LoadDataSec(m)
}

// Open - read las file
func (las *Las) Open(fileName string) (int, error) {
	var err error
	las.File, err = os.Open(fileName)
	if err != nil {
		return 0, err //FATAL error - file not exist
	}
	defer las.File.Close()
	las.FileName = fileName
	return las.Load(las.File)
}

/*LoadHeader - read las file and load all section before ~A
   returns the row number with which the data section begins, until return nil in any case
1. читаем строку
2. если коммент или пустая в игнор
3. если начало секции, определяем какой
4. если началась секция данных заканчиваем
5. читаем одну строку (это один параметер из известной нам секции)
*/
func (las *Las) LoadHeader() (int, error) {
	var (
		sec HeaderSection
	)
	//secNum := 0
	las.currentLine = 0
	for _, s := range las.rows {
		s = strings.TrimSpace(s)
		las.currentLine++
		if isIgnoredLine(s) {
			continue
		}
		if s[0] == '~' { //start new section
			if las.isDataSection(rune(s[1])) {
				break // reached the data section, stop load header
			}
			sec = las.section(rune(s[1]))
			continue
		}
		//not comment, not empty and not new section => parameter, read it
		p, w := sec.parse(s, las.currentLine)
		if !w.Empty() {
			las.addWarning(w)
		}
		p.Name = sec.uniqueName(p.Name) //TODO if a duplicate of any parameter is detected, a warning should be generated
		sec.params[p.Name] = p
		if sec.name == 'C' { //for ~Curve section need additional actions
			err := las.readCurveParam(s) //make new curve from "s" and store to container "Logs"
			if err != nil {
				las.addWarning(TWarning{directOnRead, 3, las.currentLine, fmt.Sprintf("param: '%s' error: %v", s, err)})
			}
		}
	}
	return las.currentLine, nil
}

// isDataSection - return true if data section reached
func (las *Las) isDataSection(r rune) bool {
	return (r == 0x41) || (r == 0x61)
}

// ~V - section vertion
// ~W - well info section
// ~C - curve info section
// ~A - data section
func (las *Las) section(r rune) HeaderSection {
	switch r {
	case 0x56, 0x76: //V, v
		return las.VerSec //version section
	case 0x57, 0x77: //W, w
		if las.VERS() < 2.0 {
			las.WelSec.parse = welParse12 //change parser, by default using parser 2.0
		}
		return las.WelSec //well info section
	case 0x43, 0x63: //C, c
		return las.CurSec //curve section
	case 0x50, 0x70: //P, p
		return las.ParSec //data section
	default:
		return las.OthSec
	}
}

// saveHeaderWarning - забирает и сохраняет варнинги от всех проверок
func (las *Las) storeHeaderWarning(chkResults CheckResults) {
	for _, v := range chkResults {
		las.addWarning(v.warning)
	}
}

//Разбор одной строки с мнемоникой каротажа
//Разбираем а потом сохраняем в slice
//Каждый каротаж характеризуется тремя именами
//Name     - имя каротажа, повторятся не может, если есть повторение, то Name строится добавлением к IName индекса
//IName    - имя каротажа в исходном файле, может повторятся
//Mnemonic - мнемоника, берётся из словаря, если в словаре не найдено, то ""
func (las *Las) readCurveParam(s string) error {
	l := NewLasCurve(s, las)
	las.Logs = append(las.Logs, l) //добавление в хранилище кривой каротажа с колонкой глубин
	return nil
}

// тестирование на монотонность трёх последних точек глубин внесённых в контейнер
func (las *Las) deptMonotony() CheckRes {
	if len(las.Logs) == 0 || len(las.Logs[0].D) <= 3 {
		return CheckRes{"DPTM", TWarning{directOnRead, lasSecWellInfo, las.currentLine, ""}, nil, true}
	}
	i := len(las.Logs[0].D) - 1 // индекс последней добавленной в контейнер глубины index of last element in container
	res := (las.Logs[0].D[i] - las.Logs[0].D[i-1]) == (las.Logs[0].D[i-1] - las.Logs[0].D[i-2])
	return CheckRes{"DPTM", TWarning{directOnRead, lasSecWellInfo, las.currentLine, "depth not monotony"}, nil, res}
}

// LoadDataSec - read data section from rows
func (las *Las) LoadDataSec(m int) (int, error) {
	var (
		v    float64
		err  error
		dept float64
	)
	n := len(las.Logs)                                       // количество каротажей, столько колонок данных ожидаем
	dataRows := las.rows[m:]                                 // reslice to lines with data only
	nullAsStr := strconv.FormatFloat(las.NULL(), 'f', 5, 64) // Null as string
	las.currentLine = m - 1
	for _, line := range dataRows {
		las.currentLine++
		line = strings.TrimSpace(line) // reslice
		if isIgnoredLine(line) {
			continue
		}
		fields := strings.Fields(line) //separators: tab and space
		//line must have n columns
		if len(fields) == 0 { // empty line: warning and ignore
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "wow this happened, the line is empty, ignore"})
			continue
		}
		if len(fields) != n {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("line contains %d columns, expected: %d", len(fields), n)})
		}
		// we will analyze the first column separately to check for monotony, and if occure error on parse first column then all line ignore
		dept, err = strconv.ParseFloat(fields[0], 64)
		if err != nil {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("dept:'%s' not numeric, line ignore", fields[0])})
			continue
		}
		las.Logs[0].D = append(las.Logs[0].D, dept)
		las.Logs[0].V = append(las.Logs[0].V, dept) //TODO надо подумать про колонку значений для кривой DEPT
		// проверка монотонности шага
		if cr := las.deptMonotony(); !cr.res {
			las.addWarning(cr.warning)
		}

		for j := 1; j < n; j++ { // цикл по каротажам
			s := ""
			if j >= len(fields) {
				s = nullAsStr // columns count in current line less then curves count, fill as null value
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("for column %d data not present, value set to NULL", j+1)})
			} else {
				s = fields[j]
			}
			v, err = strconv.ParseFloat(s, 64)
			if err != nil {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("error convert string: '%s' to number, set to NULL", s)})
				v = las.NULL()
			}
			las.Logs[j].D = append(las.Logs[j].D, dept)
			las.Logs[j].V = append(las.Logs[j].V, v)
		}
	}
	return las.NumPoints(), nil
}

// NumPoints - return actually number of points in data
func (las *Las) NumPoints() int {
	if len(las.Logs) == 0 {
		return 0
	}
	return len(las.Logs[0].D)
}

// Dept - return slice of DEPT curve (first column)
func (las *Las) Dept() []float64 {
	if len(las.Logs) <= 0 {
		return []float64{}
	}
	return las.Logs[0].D
}

//Save - save to file
//rewrite if file exist
//if useMnemonic == true then on save using std mnemonic on ~Curve section
//TODO las have field filename of readed las file, after save filename must update or not? warning occure on write for what file?
func (las *Las) Save(fileName string, useMnemonic ...bool) error {
	var (
		err       error
		bufToSave []byte
	)
	if len(useMnemonic) > 0 {
		bufToSave, err = las.SaveToBuf(useMnemonic[0]) //TODO las.SaveToBuf(true) not test
	} else {
		bufToSave, err = las.SaveToBuf(false)
	}
	if err != nil {
		return err //TODO случай возврата ошибки из las.SaveToBuf() не тестируется
	}
	if !xlib.FileExists(fileName) {
		err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
		if err != nil {
			return errors.New("path: '" + filepath.Dir(fileName) + "' can't create >>" + err.Error())
		}
	}
	err = ioutil.WriteFile(fileName, bufToSave, 0644)
	if err != nil {
		return errors.New("file: '" + fileName + "' can't open to write >>" + err.Error())
	}
	return nil
}

// SaveToBuf - save to file
// rewrite if file exist
// if useMnemonic == true then on save using std mnemonic on ~Curve section
// ir return err != nil then fatal error, returned slice is not full corrected
func (las *Las) SaveToBuf(useMnemonic bool) ([]byte, error) {
	n := len(las.Logs) //log count
	if n <= 0 {
		return nil, errors.New("logs not exist")
	}
	var b bytes.Buffer
	fmt.Fprint(&b, _LasFirstLine)
	fmt.Fprintf(&b, _LasVersion, 2.0) //file is always saved in 2.0 format
	fmt.Fprint(&b, _LasWrap)
	fmt.Fprint(&b, _LasWellInfoSec)
	fmt.Fprintf(&b, _LasStrt, las.STRT())
	fmt.Fprintf(&b, _LasStop, las.STOP())
	fmt.Fprintf(&b, _LasStep, las.STEP())
	fmt.Fprintf(&b, _LasNull, las.NULL())
	fmt.Fprintf(&b, _LasWell, las.WELL())
	fmt.Fprint(&b, _LasCurvSec)
	fmt.Fprint(&b, _LasCurvDept)

	for i := 1; i < n; i++ { //Пишем названия каротажей
		l := las.Logs[i]
		if useMnemonic {
			if len(l.Mnemonic) > 0 {
				l.Name = l.Mnemonic
			}
		}
		fmt.Fprintf(&b, _LasCurvLine, l.Name, l.Unit) //запись мнемоник в секции ~Curve
	}
	fmt.Fprint(&b, _LasDataSec)
	fmt.Fprintf(&b, "%s\n", las.Logs.Captions()) //write comment with curves name
	//write data
	for i := 0; i < las.NumPoints(); i++ { //loop by dept (.)
		fmt.Fprintf(&b, "%-10.4f ", las.Logs[0].D[i])
		for j := 1; j < n; j++ { //loop by logs
			fmt.Fprintf(&b, "%-10.4f ", las.Logs[j].V[i])
		}
		fmt.Fprintln(&b)
	}
	r, _ := cpd.NewReaderTo(io.Reader(&b), las.oCodepage.String()) //ошибку не обрабатываем, допустимость oCodepage проверяем раньше, других причин нет
	bufToSave, _ := ioutil.ReadAll(r)
	return bufToSave, nil
}

// IsEmpty - test to not initialize object
func (las *Las) IsEmpty() bool {
	return (las.Logs == nil)
}

// GetStrtFromData - return strt from data section
// read 1 line from section ~A and determine strt
// close file
// return Null if error occurs
func (las *Las) GetStrtFromData() float64 {
	iFile, err := os.Open(las.FileName)
	if err != nil {
		return las.NULL() //не обрабатывается в тесте
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileStop(las.FileName, "~A")
	if (err != nil) || (iScanner == nil) {
		return las.NULL() //не обрабатывается в тесте
	}

	s := ""
	dept1 := 0.0
	for i := 0; iScanner.Scan(); i++ {
		s = strings.TrimSpace(iScanner.Text())
		if (len(s) == 0) || (s[0] == '#') {
			continue //не обрабатывается в тесте
		}
		k := strings.IndexRune(s, ' ')
		if k < 0 {
			k = len(s)
		}
		dept1, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			return las.NULL() //не обрабатывается в тесте
		}
		return dept1
	}
	//если мы попали сюда, то всё грусно, в файле после ~A не нашлось двух строчек с данными... или пустые строчки или комменты
	return las.NULL()
}

// GetStepFromData - return step from data section
// read 2 line from section ~A and determine step
// close file
// return Null if error occure
func (las *Las) GetStepFromData() float64 {
	iFile, err := os.Open(las.FileName)
	if err != nil {
		return las.NULL() //не обрабатывается в тесте
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileStop(las.FileName, "~A")
	if (err != nil) || (iScanner == nil) {
		return las.NULL()
	}

	s := ""
	j := 0
	dept1 := 0.0
	dept2 := 0.0
	for i := 0; iScanner.Scan(); i++ {
		s = strings.TrimSpace(iScanner.Text())
		if isIgnoredLine(s) {
			continue
		}
		k := strings.IndexRune(s, ' ')
		if k < 0 {
			k = len(s)
		}
		dept1, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			// case if the data row in the first position (dept place) contains not a number
			return las.NULL()
		}
		j++
		if j == 2 {
			// good case, found two points and determined the step
			return math.Round((dept1-dept2)*10) / 10
		}
		dept2 = dept1
	}
	//bad case, data section not contain two rows with depth
	return las.NULL() //не обрабатывается в тесте
}

func (las *Las) setStep(h float64) {
	las.WelSec.params["STEP"] = HeaderParam{strconv.FormatFloat(h, 'f', -1, 64), "STEP", "", "", "", "step of index", 8}
}

func (las *Las) setStrt(strt float64) {
	las.WelSec.params["STRT"] = HeaderParam{strconv.FormatFloat(strt, 'f', -1, 64), "STRT", "", "", "", "first index value", 6}
}

// IsStrtEmpty - return true if parameter Strt not exist in file
func (las *Las) IsStrtEmpty() bool {
	return las.STRT() == StdNull
}

// IsStopEmpty - return true if parameter Stop not exist in file
func (las *Las) IsStopEmpty() bool {
	return las.STOP() == StdNull
}

// IsStepEmpty - return true if parameter Step not exist in file
func (las *Las) IsStepEmpty() bool {
	return las.STEP() == StdNull
}

// SetNull - change parameter NULL in WELL INFO section and in all logs
func (las *Las) SetNull(null float64) {
	for _, l := range las.Logs { //loop by logs
		for i := range l.V { //loop by dept step
			if l.V[i] == las.NULL() {
				l.V[i] = null
			}
		}
	}
	las.WelSec.params["NULL"] = HeaderParam{strconv.FormatFloat(null, 'f', -1, 64), "NULL", "", "", "", "null value", las.WelSec.params["NULL"].lineNo}
}

// SaveWarning - save to file all warning
func (las *Las) SaveWarning(fileName string) error {
	if las.Warnings.Count() == 0 {
		return nil
	}
	oFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	las.SaveWarningToFile(oFile)
	oFile.Close()
	return nil
}

// SaveWarningToWriter - store all warning to writer, return count lines writed to
func (las *Las) SaveWarningToWriter(writer *bufio.Writer) int {
	n := las.Warnings.Count()
	if n == 0 {
		return 0
	}
	for _, w := range las.Warnings {
		writer.WriteString(w.String())
		writer.WriteString("\n")
	}
	return n
}

// SaveWarningToFile - store all warning to file, file not close. return count warning writed
func (las *Las) SaveWarningToFile(oFile *os.File) int {
	if oFile == nil {
		return 0
	}
	if las.Warnings.Count() == 0 {
		return 0
	}
	oFile.WriteString("**file: " + las.FileName + "**\n")
	n := las.Warnings.SaveWarningToFile(oFile)
	oFile.WriteString("\n")
	return n
}

func (las *Las) addWarning(w TWarning) {
	if las.Warnings.Count() < las.maxWarningCount {
		las.Warnings = append(las.Warnings, w)
		if las.Warnings.Count() == las.maxWarningCount {
			las.Warnings = append(las.Warnings, TWarning{0, 0, -1, "*maximum count* of warning reached, change parameter 'maxWarningCount' in 'glas.ini'"})
		}
	}
}

// GetMnemonic - return Mnemonic from dictionary by Log Name,
// if Mnemonic not found return ""
// if Dictionary is nil, then return ""
func (las *Las) GetMnemonic(logName string) string {
	if (las.LogDic == nil) || (las.VocDic == nil) {
		return "" //"-"
	}
	_, ok := (*las.LogDic)[logName]
	if ok { //GOOD - название каротажа равно мнемонике
		return logName
	}
	v, ok := (*las.VocDic)[logName]
	if ok { //POOR - название загружаемого каротажа найдено в словаре подстановок, мнемоника найдена
		return v
	}
	return ""
}
