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
// at save file code page converted to specifyed in Las.toCodePage
//TODO add pointer to cfg
//TODO при создании объекта las есть возможность указать кодировку записи, нужна возможность указать явно кодировку чтения
type Las struct {
	rows            []string           // buffer for read source file, only converted to UTF-8 no any othe change
	nRows           int                // actually count of lines in source file
	FileName        string             // file name from load
	File            *os.File           // the file from which we are reading
	Reader          io.Reader          // reader created from File, provides decode from codepage to UTF-8
	scanner         *bufio.Scanner     // scanner
	Ver             float64            // version 1.0, 1.2 or 2.0
	Wrap            string             // YES || NO
	Strt            float64            // start depth
	Stop            float64            // stop depth
	Step            float64            // depth step
	Null            float64            // value interpreted as empty
	Well            string             // well name
	Rkb             float64            // altitude KB
	Logs            LasCurves          // store all logs
	LogDic          *map[string]string // external dictionary of standart log name - mnemonics
	VocDic          *map[string]string // external vocabulary dictionary of log mnemonic
	Warnings        TLasWarnings       // slice of warnings occure on read or write
	oCodepage       cpd.IDCodePage     // codepage to save, default xlib.CpWindows1251. to special value, specify at make: NewLas(cp...)
	iDuplicate      int                // index of duplicated mnemonic, increase by 1 if found duplicated
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

func (las *Las) parFloat(name string, defValue float64) float64 {
	v := defValue
	if p, ok := las.VerSec.params[name]; ok {
		v, _ = strconv.ParseFloat(p.Val, 64)
	}
	return v
}

func (las *Las) parStr(name, defValue string) string {
	v := defValue
	if p, ok := las.VerSec.params[name]; ok {
		v = p.Val
	}
	return v
}

// NULL - return null value of las file as float64
// if parameter NULL in las file not exist, then return StdNull (by default -999.25)
func (las *Las) NULL() float64 {
	return las.parFloat("NULL", StdNull)
}

// STOP - return depth stop value of las file as float64
// if parameter STOP in las file not exist, then return StdNull (by default -999.25)
func (las *Las) STOP() float64 {
	return las.parFloat("STOP", StdNull)
}

// STRT - return depth start value of las file as float64
// if parameter NULL in las file not exist, then return StdNull (by default -999.25)
func (las *Las) STRT() float64 {
	return las.parFloat("STRT", StdNull)
}

// STEP - return depth step value of las file as float64
// if parameter not exist, then return StdNull (by default -999.25)
func (las *Las) STEP() float64 {
	return las.parFloat("STEP", StdNull)
}

// VERS - return version of las file as float64
// if parameter VERS in las file not exist, then return 2.0
func (las *Las) VERS() float64 {
	return las.parFloat("VERS", 2.0)
}

// WRAP - return wrap parameter of las file
// if parameter not exist, then return "NO"
func (las *Las) WRAP() string {
	return las.parStr("WRAP", "NO")
}

// WELL - return well name
// if parameter WELL in las file not exist, then return "--"
func (las *Las) WELL() string {
	return las.parStr("WELL", "--")
}

//NewLas - make new object Las class
//autodetect code page at load file
//code page to save by default is cpd.CP1251
func NewLas(outputCP ...cpd.IDCodePage) *Las {
	las := &Las{} //new(Las)
	las.Ver = 2.0 //<VER>
	las.Wrap = "NO"
	las.nRows = ExpPoints
	las.rows = make([]string, 0, las.nRows)
	las.Logs = make([]LasCurve, 0)
	las.VerSec = NewVerSection()
	las.WelSec = NewWelSection()
	las.CurSec = NewCurSection()
	las.ParSec = NewParSection()
	las.OthSec = NewOthSection()
	las.maxWarningCount = MaxWarningCount
	las.stdNull = StdNull
	las.Strt = StdNull
	las.Stop = StdNull
	las.Step = StdNull
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

// IsWraped - return true if WRAP == YES
func (las *Las) IsWraped() bool {
	return strings.Contains(strings.ToUpper(las.Wrap), "Y")
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

// Open -
func (las *Las) Open(fileName string) (int, error) {
	var err error
	las.File, err = os.Open(fileName)
	if err != nil {
		return 0, err //FATAL error - file not exist
	}
	defer las.File.Close()
	las.FileName = fileName
	//create Reader, this reader decode to UTF-8
	las.Reader, err = cpd.NewReader(las.File)
	if err != nil {
		return 0, err //FATAL error - file cannot be decoded to UTF-8
	}
	// prepare file to read
	las.scanner = bufio.NewScanner(las.Reader)
	// ePoints - содержит теперь количество строк в las файле
	// соответственно слайсы для хранения кривых будут создаваться этой длины
	//las.ePoints = las.ReadRows()
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
		if h == las.Null {
			las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprint("__WRN__ STRT parameter on data is wrong setting to 0")})
			las.setStrt(0)
		}
		las.setStrt(h)
	}
	if r.stepWrong() {
		h := las.GetStepFromData() // return las.Null if cannot calculate step from data
		if h == las.Null {
			las.addWarning(TWarning{directOnRead, lasSecWellInfo, las.currentLine, fmt.Sprint("__WRN__ STEP parameter on data is wrong")})
		}
		las.setStep(h)
	}
	return las.LoadDataSec(m)
}

/*LoadHeader - read las file and load all section before ~A
   returns the row number with which the data section begins, until return nil in any case
   secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - A data
1. читаем строку
2. если коммент или пустая в игнор
3. если начало секции, определяем какой
4. если началась секция данных заканчиваем
5. читаем одну строку (это один параметер из известной нам секции)
*/
func (las *Las) LoadHeader() (int, error) {
	var (
		err error
		sec HeaderSection
	)
	secNum := 0
	las.currentLine = 0
	for i, s := range las.rows {
		s = strings.TrimSpace(s)
		las.currentLine++
		if isIgnoredLine(s) {
			continue
		}
		if s[0] == '~' { //start new section
			secNum = las.selectSection(rune(s[1]))
			if secNum == lasSecData {
				break // reached the data section, stop load header
			}
			sec = las.section(rune(s[1]))
			continue
		}
		//not comment, not empty and not new section => parameter, read it
		err = las.ReadParameter(s, secNum)
		p, _ := sec.parse(s, las.currentLine)
		p.Name = sec.uniqueName(p.Name)
		sec.params[p.Name] = p
		if err != nil {
			las.addWarning(TWarning{directOnRead, secNum, i, fmt.Sprintf("param: '%s' error: %v", s, err)})
		}
	}
	return las.currentLine, nil
}

// selectSection - analize first char after ~
// ~V - section vertion
// ~W - well info section
// ~C - curve info section
// ~A - data section
func (las *Las) selectSection(r rune) int {
	switch r {
	case 0x76, 0x56: //86, 118: //V, v
		return lasSecVersion //version section
	case 0x77, 0x57: //W, w
		return lasSecWellInfo //well info section
	case 0x43, 0x63: //C, c
		return lasSecCurInfo //curve section
	case 0x41, 0x61: //A, a
		return lasSecData //data section
	default:
		return lasSecIgnore
	}
}

func (las *Las) section(r rune) HeaderSection {
	switch r {
	case 0x56, 0x76: //V, v
		return las.VerSec //version section
	case 0x57, 0x77: //W, w
		if las.Ver < 2.0 {
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

// ReadParameter - read one parameter
func (las *Las) ReadParameter(s string, secNum int) error {
	switch secNum {
	case lasSecVersion:
		return las.readVersionParam(s)
	case lasSecWellInfo:
		return las.ReadWellParam(s)
	case lasSecCurInfo:
		return las.readCurveParam(s)
	}
	return nil
}

func (las *Las) readVersionParam(s string) error {
	var err error
	p := NewHeaderParam(s, 0)
	switch p.Name {
	case "VERS":
		las.Ver, err = strconv.ParseFloat(p.Val, 64)
	case "WRAP":
		las.Wrap = p.Val
	}
	return err
}

//ReadWellParam - read parameter from WELL section
func (las *Las) ReadWellParam(s string) error {
	var err error
	p := NewHeaderParam(s, 0)
	switch p.Name {
	case "STRT":
		las.Strt, err = strconv.ParseFloat(p.Val, 64)
	case "STOP":
		las.Stop, err = strconv.ParseFloat(p.Val, 64)
	case "STEP":
		las.Step, err = strconv.ParseFloat(p.Val, 64)
	case "NULL":
		las.Null, err = strconv.ParseFloat(p.Val, 64)
	case "WELL":
		if las.Ver < 2.0 {
			las.Well = p.Desc
		} else {
			las.Well = wellNameFromParam(p)
		}
	}
	if err != nil {
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("detected param: %v, unit:%v, value: %v\n", p.Name, p.Unit, p.Val)})
	}
	return err
}

//Разбор одной строки с мнемоникой каротажа
//Разбираем а потом сохраняем в slice
//Каждый каротаж характеризуется тремя именами
//IName    - имя каротажа в исходном файле, может повторятся
//Name     - ключ в map хранилище, повторятся не может. если в исходном есть повторение, то Name строится добавлением к IName индекса
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
	n := len(las.Logs)                                     // количество каротажей, столько колонок данных ожидаем
	dataRows := las.rows[m:]                               // reslice to lines with data only
	nullAsStr := strconv.FormatFloat(las.Null, 'f', 5, 64) // Null as string
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
				v = las.Null
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
	fmt.Fprintf(&b, _LasStrt, las.Strt)
	fmt.Fprintf(&b, _LasStop, las.Stop)
	fmt.Fprintf(&b, _LasStep, las.Step)
	fmt.Fprintf(&b, _LasNull, las.Null)
	fmt.Fprintf(&b, _LasWell, las.Well)
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
		return las.Null //не обрабатывается в тесте
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileStop(las.FileName, "~A")
	if (err != nil) || (iScanner == nil) {
		return las.Null //не обрабатывается в тесте
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
			return las.Null //не обрабатывается в тесте
		}
		return dept1
	}
	//если мы попали сюда, то всё грусно, в файле после ~A не нашлось двух строчек с данными... или пустые строчки или комменты
	return las.Null
}

// GetStepFromData - return step from data section
// read 2 line from section ~A and determine step
// close file
// return Null if error occure
func (las *Las) GetStepFromData() float64 {
	iFile, err := os.Open(las.FileName)
	if err != nil {
		return las.Null //не обрабатывается в тесте
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileStop(las.FileName, "~A")
	if (err != nil) || (iScanner == nil) {
		return las.Null
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
			return las.Null
		}
		j++
		if j == 2 {
			// good case, found two points and determined the step
			return math.Round((dept1-dept2)*10) / 10
		}
		dept2 = dept1
	}
	//bad case, data section not contain two rows with depth
	return las.Null //не обрабатывается в тесте
}

func (las *Las) setStep(h float64) {
	las.Step = h
}

func (las *Las) setStrt(strt float64) {
	las.Strt = strt
}

// IsStrtEmpty - return true if parameter Strt not exist in file
func (las *Las) IsStrtEmpty() bool {
	return las.Strt == StdNull
}

// IsStopEmpty - return true if parameter Stop not exist in file
func (las *Las) IsStopEmpty() bool {
	return las.Stop == StdNull
}

// IsStepEmpty - return true if parameter Step not exist in file
func (las *Las) IsStepEmpty() bool {
	return las.Step == StdNull
}

// SetNull - change parameter NULL in WELL INFO section and in all logs
func (las *Las) SetNull(aNull float64) {
	for _, l := range las.Logs { //loop by logs
		for i := range l.V { //loop by dept step
			if l.V[i] == las.Null {
				l.V[i] = aNull
			}
		}
	}
	las.Null = aNull
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
