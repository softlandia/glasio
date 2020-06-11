// (c) softland 2020
// softlandia@gmail.com

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

///format strings represent structure of LAS file
const (
	_LasFirstLine      = "~Version information\n"
	_LasVersion        = "VERS.                          %3.1f : glas (c) softlandia@gmail.com\n"
	_LasWrap           = "WRAP.                          NO  : ONE LINE PER DEPTH STEP\n"
	_LasWellInfoSec    = "~Well information\n"
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
	_LasDataSec        = "~ASCII Log Data\n"

	//secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - dAta
	lasSecIgnore   = 0
	lasSecVersion  = 1
	lasSecWellInfo = 2
	lasSecCurInfo  = 3
	lasSecData     = 4
)

// HeaderParam - any parameter of LAS
type HeaderParam struct {
	lineNo  int // number of line in source file
	source, // text line from source file contain this parameter
	name, // name of parameter: STOP, WELL, SP - curve name also
	Val, // parameter value
	Unit, // unit of parameter
	Description string // descrioption of parameter
}

// HeaderSection - contain parameters of Well section
type HeaderSection map[string]HeaderParam

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
	ePoints         int                // expected count (.)
	nPoints         int                // actually count (.)
	oCodepage       cpd.IDCodePage     // codepage to save, default xlib.CpWindows1251. to special value, specify at make: NewLas(cp...)
	iDuplicate      int                // index of duplicated mnemonic, increase by 1 if found duplicated
	currentLine     int                // index of current line in readed file
	maxWarningCount int                // default maximum warning count
	stdNull         float64            // default null value
	VerSec,
	WellSec,
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

//NewLas - make new object Las class
//autodetect code page at load file
//code page to save by default is cpd.CP1251
func NewLas(outputCP ...cpd.IDCodePage) *Las {
	las := new(Las)
	las.Ver = 2.0
	las.Wrap = "NO"
	// до того как прочитаем данные мы не можем знать сколько их фактически, предполагаем что 1000, достаточно большой буфер
	// избавит от необходимости довыделять память при чтении
	las.ePoints = ExpPoints
	las.nRows = ExpPoints
	las.rows = make([]string, 0, las.nRows)
	las.Logs = make(map[string]LasCurve)
	las.VerSec = make(HeaderSection)
	las.WellSec = make(HeaderSection)
	las.CurSec = make(HeaderSection)
	las.ParSec = make(HeaderSection)
	las.OthSec = make(HeaderSection)
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

// selectSection - analize first char after ~
// ~V - section vertion
// ~W - well info section
// ~C - curve info section
// ~A - data section
func (las *Las) selectSection(r rune) int {
	switch r {
	case 86: //V
		return lasSecVersion //version section
	case 118: //v
		return lasSecVersion //version section
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

// IsWraped - return true if WRAP == YES
func (las *Las) IsWraped() bool {
	return strings.Contains(strings.ToUpper(las.Wrap), "Y") //(strings.Index(strings.ToUpper(o.Wrap), "Y") >= 0)
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
			las.Warnings = append(las.Warnings, TWarning{0, 0, 0, "*maximum count* of warning reached, change parameter 'maxWarningCount' in 'glas.ini'"})
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

// Open - load las file
// return error on:
// - file not exist
// - file cannot be decoded to UTF-8
// - las is wrapped
// - las file not contain Curve section
func (las *Las) Open(fileName string) (int, error) {
	var err error
	las.File, err = os.Open(fileName)
	if err != nil {
		return 0, err //FATAL error - file not exist
	}
	defer las.File.Close()
	las.FileName = fileName
	//create and store Reader, this reader decode to UTF-8
	las.Reader, err = cpd.NewReader(las.File)
	if err != nil {
		return 0, err //FATAL error - file cannot be decoded to UTF-8
	}
	// prepare file to read
	las.scanner = bufio.NewScanner(las.Reader)
	las.currentLine = 0
	las.LoadHeader()
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
	return las.ReadDataSec(fileName)
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

// Open2 -
func (las *Las) Open2(fileName string) (int, error) {
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
	las.ePoints = las.ReadRows()
	//las.currentLine = 0
	m, _ := las.LoadHeader2()
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
	return m, nil //las.ReadDataSec2(m)
}

// LoadHeader2 - new version, after test will be replace verion LoadHeader
// returns the row number with which the data section begins, contain ~A
func (las *Las) LoadHeader2() (int, error) {
	s := ""
	var err error
	secNum := 0
	for i := 0; i < len(las.rows); i++ {
		s = strings.TrimSpace(las.rows[i])
		las.currentLine++
		if isIgnoredLine(s) {
			continue
		}
		if s[0] == '~' { //start new section
			secNum = las.selectSection(rune(s[1]))
			if secNum == lasSecData {
				break // reached the data section, stop load header
			}
		} else {
			//if not comment, not empty and not new section => parameter, read it
			err = las.ReadParameter(s, secNum)
			if err != nil {
				las.addWarning(TWarning{directOnRead, secNum, i, fmt.Sprintf("param: '%s' error: %v", s, err)})
			}
		}
	}
	return las.currentLine, nil
}

// saveHeaderWarning - забирает и сохраняет варнинги от всех проверок
func (las *Las) storeHeaderWarning(chkResults CheckResults) {
	for _, v := range chkResults {
		las.addWarning(v.warning)
	}
}

/*LoadHeader - read las file and load all section before ~A
   secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - A data
1. читаем строку
2. если коммент или пустая в игнор
3. если начало секции, определяем какой
4. если началась секция данных заканчиваем
5. читаем одну строку (это один параметер из известной нам секции)
   Пока всегда возвращает nil, причин возвращать другое значение пока нет.
*/
func (las *Las) LoadHeader() error {
	s := ""
	var err error
	secNum := 0
	for i := 0; las.scanner.Scan(); i++ {
		s = strings.TrimSpace(las.scanner.Text())
		las.currentLine++
		if isIgnoredLine(s) {
			continue
		}
		if s[0] == '~' { //start new section
			secNum = las.selectSection(rune(s[1]))
			if secNum == lasSecData {
				break // reached the data section, stop load header
			}
		} else {
			//if not comment, not empty and not new section => parameter, read it
			err = las.ReadParameter(s, secNum)
			if err != nil {
				las.addWarning(TWarning{directOnRead, secNum, las.currentLine, fmt.Sprintf("param: '%s' error: %v", s, err)})
			}
		}
	}
	return nil
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
	p := NewLasParam(s)
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
	p := NewLasParam(s)
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

//ChangeDuplicateLogName - return non duplicated name of log
//if input name unique, return input name
//if input name not unique, return input name + index duplicate
//index duplicate - Las field, increase
func (las *Las) ChangeDuplicateLogName(name string) string {
	s := ""
	if _, ok := las.Logs[name]; ok {
		las.iDuplicate++
		s = fmt.Sprintf("%v", las.iDuplicate)
		name += s
	}
	return name
}

//Разбор одной строки с мнемоникой каротажа
//Разбираем в переменную l а потом сохраняем в map
//Каждый каротаж характеризуется тремя именами
//IName    - имя каротажа в исходном файле, может повторятся
//Name     - ключ в map хранилище, повторятся не может. если в исходном есть повторение, то Name строится добавлением к IName индекса
//Mnemonic - мнемоника, берётся из словаря, если в словаре не найдено, то ""
func (las *Las) readCurveParam(s string) error {
	l := NewLasCurve(s)
	// поскольку этот метод вызывается для каждой кривой в секции ~Curve, то заново определять для каждой кривой количество ожидаемых
	// точек через las.GetExpectedPointsCount() СТРАННО
	l.Init(len(las.Logs), las.GetMnemonic(l.Name), las.ChangeDuplicateLogName(l.Name), las.GetExpectedPointsCount(las.Strt, las.Stop, las.Step))
	las.Logs[l.Name] = l //добавление в хранилище кривой каротажа с колонкой глубин
	return nil
}

// GetExpectedPointsCount - оценка количества точек по параметрам STEP, STRT, STOP
// перед чтением данных пытаемся оценить их количество
// в правильной ситуации количество должно быть равныи (stop - strt)/step
// случаи ошибок:
// 1. step == 0
// 2. strt == stop
// при ошибке вернём предполагаемое (заданное по умолчанию) количество
// стандартные случаи:
// 1. strt < stop & step > 0
// 2. stop < strt & step < 0
func (las *Las) GetExpectedPointsCount(strt, stop, step float64) (m int) {
	if step == 0.0 {
		return las.ePoints
	}
	d := stop - strt
	if d == 0 {
		return las.ePoints
	}
	m = int(d/step) + 2
	if m < 0 {
		m = -m
	}
	return m
}

//expandDept - if actually data points exceeds
func (las *Las) expandDept(d *LasCurve) {
	//actual number of points more then expected
	las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "actual number of data lines more than expected, check: STRT, STOP, STEP"})
	las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "expand number of points"})
	//ожидаем удвоения данных
	las.ePoints *= 2
	//expand first log - dept
	newDept := make([]float64, las.ePoints)
	copy(newDept, d.D)
	d.D = newDept

	newLog := make([]float64, las.ePoints)
	copy(newLog, d.D)
	d.V = newLog
	las.Logs[d.Name] = *d

	//loop over other logs
	n := len(las.Logs)
	var l *LasCurve
	for j := 1; j < n; j++ {
		l, _ = las.logByIndex(j)
		newDept := make([]float64, las.ePoints)
		copy(newDept, l.D)
		l.D = newDept

		newLog := make([]float64, las.ePoints)
		copy(newLog, l.V)
		l.V = newLog
		las.Logs[l.Name] = *l
	}
}

func isSeparator(r rune) bool {
	switch r {
	case 9, 32:
		return true
	default:
		return false
	}
}

// ReadDataSec - read section of data
func (las *Las) ReadDataSec(fileName string) (int, error) {
	var (
		v    float64
		err  error
		d    *LasCurve
		l    *LasCurve
		dept float64
		i    int
	)

	// исходя из параметров STRT, STOP и STEP определяем ожидаемое количество строк данных
	// данную операцию уже выполняли много раз при считывании кривых
	las.ePoints = las.GetExpectedPointsCount(las.Strt, las.Stop, las.Step)
	n := len(las.Logs)       //количество каротажей, столько колонок данных ожидаем
	d, _ = las.logByIndex(0) //dept log
	s := ""
	for i = 0; las.scanner.Scan(); i++ {
		las.currentLine++
		if i == las.ePoints {
			las.expandDept(d)
		}
		s = strings.TrimSpace(las.scanner.Text())
		// i счётчик не считанных строк, а фактически считанных данных - счётчик добавлений в слайсы данных
		//TODO возможно следует завести отдельный счётчик и оставить в покое счётчик цикла
		if isIgnoredLine(s) {
			i--
			continue
		}
		//first column is DEPT
		k := strings.IndexFunc(s, isSeparator)
		if k < 0 { //line must have n+1 column and n separated spaces block (+1 becouse first column DEPT)
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("line: %d is empty, ignore", las.currentLine)})
			i--
			continue
		}
		dept, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("first column '%s' not numeric, ignore", s[:k])})
			i--
			continue
		}
		d.D[i] = dept
		// проверка шага у первых двух точек данных и сравнение с параметром step
		//TODO данную проверку следует делать через Checker
		if i > 1 {
			if math.Pow(((dept-d.D[i-1])-las.Step), 2) > 0.1 {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("actual step %5.2f ≠ global STEP %5.2f", (dept - d.D[i-1]), las.Step)})
			}
		}
		// проверка шага между точками [i-1, i] и точками [i-2, i-1] обнаружение немонотонности колонки глубин
		if i > 2 {
			if math.Pow(((dept-d.D[i-1])-(d.D[i-1]-d.D[i-2])), 2) > 0.1 {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("step %5.2f ≠ previously step %5.2f", (dept - d.D[i-1]), (d.D[i-1] - d.D[i-2]))})
				dept = d.D[i-1] + las.Step
			}
		}

		s = strings.TrimSpace(s[k+1:]) //cut first column
		//цикл по каротажам
		for j := 1; j < (n - 1); j++ {
			iSpace := strings.IndexFunc(s, isSeparator)
			switch iSpace {
			case -1: //не все колонки прочитаны, а разделителей уже нет... пробуем игнорировать сроку заполняя оставшиеся каротажи NULLами
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "not all column are read, set log value to NULL"})
			case 0:
				v = las.Null
			case 1:
				v, err = strconv.ParseFloat(s[:1], 64)
			default:
				v, err = strconv.ParseFloat(s[:iSpace], 64)
			}
			if err != nil {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("can't convert string: '%s' to number, set to NULL", s[:iSpace-1])})
				v = las.Null
			}
			l, err = las.logByIndex(j)
			if err != nil {
				las.nPoints = i
				return i, errors.New("internal ERROR, func (las *Las) readDataSec()::las.logByIndex(j) return error")
			}
			l.D[i] = dept
			l.V[i] = v
			s = strings.TrimSpace(s[iSpace+1:])
		}
		//остаток - последняя колонка
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "not all column are read, set log value to NULL"})
			v = las.Null
		}
		l, err = las.logByIndex(n - 1)
		if err != nil {
			las.nPoints = i
			return i, errors.New("internal ERROR, func (las *Las) readDataSec()::las.logByIndex(j) return error on last column")
		}
		l.D[i] = dept
		l.V[i] = v
	}
	//i - actually readed lines and add (.) to data array
	//crop logs to actually len
	err = las.SetActuallyNumberPoints(i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// ReadDataSec2 - read from rows
func (las *Las) ReadDataSec2(m int) (int, error) {
	var (
		v    float64
		err  error
		d    *LasCurve
		l    *LasCurve
		dept float64
		i    int
	)
	las.ePoints = len(las.rows)
	n := len(las.Logs)       //количество каротажей, столько колонок данных ожидаем
	d, _ = las.logByIndex(0) //dept log
	s := ""
	for i = m; i < len(las.rows); i++ {
		s = strings.TrimSpace(las.rows[i])
		if isIgnoredLine(s) {
			continue
		}
		//first column is DEPT
		k := strings.IndexFunc(s, isSeparator)
		if k < 0 { //line must have n+1 column and n separated spaces block (+1 becouse first column DEPT)
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("line: %d is empty, ignore", las.currentLine)})
			continue
		}
		dept, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("first column '%s' not numeric, ignore", s[:k])})
			continue
		}
		d.D[i] = dept
		// проверка шага у первых двух точек данных и сравнение с параметром step
		//TODO данную проверку следует делать через Checker
		if i > 1 {
			if math.Pow(((dept-d.D[i-1])-las.Step), 2) > 0.1 {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("actual step %5.2f ≠ global STEP %5.2f", (dept - d.D[i-1]), las.Step)})
			}
		}
		// проверка шага между точками [i-1, i] и точками [i-2, i-1] обнаружение немонотонности колонки глубин
		if i > 2 {
			if math.Pow(((dept-d.D[i-1])-(d.D[i-1]-d.D[i-2])), 2) > 0.1 {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("step %5.2f ≠ previously step %5.2f", (dept - d.D[i-1]), (d.D[i-1] - d.D[i-2]))})
				dept = d.D[i-1] + las.Step
			}
		}

		s = strings.TrimSpace(s[k+1:]) //cut first column
		//цикл по каротажам
		for j := 1; j < (n - 1); j++ {
			iSpace := strings.IndexFunc(s, isSeparator)
			switch iSpace {
			case -1: //не все колонки прочитаны, а разделителей уже нет... пробуем игнорировать сроку заполняя оставшиеся каротажи NULLами
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "not all column are read, set log value to NULL"})
			case 0:
				v = las.Null
			case 1:
				v, err = strconv.ParseFloat(s[:1], 64)
			default:
				v, err = strconv.ParseFloat(s[:iSpace], 64)
			}
			if err != nil {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("can't convert string: '%s' to number, set to NULL", s[:iSpace-1])})
				v = las.Null
			}
			l, err = las.logByIndex(j)
			if err != nil {
				las.nPoints = i
				return i, errors.New("internal ERROR, func (las *Las) readDataSec()::las.logByIndex(j) return error")
			}
			l.D[i] = dept
			l.V[i] = v
			s = strings.TrimSpace(s[iSpace+1:])
		}
		//остаток - последняя колонка
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "not all column are read, set log value to NULL"})
			v = las.Null
		}
		l, err = las.logByIndex(n - 1)
		if err != nil {
			las.nPoints = i
			return i, errors.New("internal ERROR, func (las *Las) readDataSec()::las.logByIndex(j) return error on last column")
		}
		l.D[i] = dept
		l.V[i] = v
	}
	//i - actually readed lines and add (.) to data array
	//crop logs to actually len
	err = las.SetActuallyNumberPoints(i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// NumPoints - return actually number of points in data
func (las *Las) NumPoints() int {
	return las.nPoints
}

// Dept - return slice of DEPT curve (first column)
func (las *Las) Dept() []float64 {
	d, err := las.logByIndex(0)
	if err != nil {
		return nil
	}
	return d.D
}

// SetActuallyNumberPoints - crop all curve to numPoints count of data
func (las *Las) SetActuallyNumberPoints(numPoints int) error {
	if numPoints <= 0 {
		las.nPoints = 0
		return errors.New("internal ERROR, func (las *Las) setActuallyNumberPoints(), actually number of points <= 0")
	}
	if numPoints > len(las.Dept()) {
		las.nPoints = 0
		return errors.New("internal ERROR, func (las *Las) setActuallyNumberPoints(), actually number of points > then exist data")
	}
	for _, l := range las.Logs {
		l.SetLen(numPoints)
	}
	las.nPoints = numPoints
	return nil
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
		bufToSave, err = las.SaveToBuf(useMnemonic[0]) //<> bufToSave, err = las.SaveToBuf(true) //TODO las.SaveToBuf(true) not test
	} else {
		bufToSave, err = las.SaveToBuf(false)
	}
	if err != nil {
		return err
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
	var err error
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

	var sb strings.Builder
	sb.WriteString("# DEPT  |") //готовим строчку с названиями каротажей глубина всегда присутствует
	var l *LasCurve
	for i := 1; i < n; i++ { //Пишем названия каротажей
		l, _ := las.logByIndex(i)
		if useMnemonic {
			if len(l.Mnemonic) > 0 {
				l.Name = l.Mnemonic
			}
		}
		fmt.Fprintf(&b, _LasCurvLine, l.Name, l.Unit) //запись мнемоник в секции ~Curve
		sb.WriteString(" ")
		fmt.Fprintf(&sb, "%-8s|", l.Name) //Собираем строчку с названиями каротажей
	}

	fmt.Fprint(&b, _LasDataSec)
	//write data
	fmt.Fprintf(&b, "%s\n", sb.String())
	dept, _ := las.logByIndex(0)
	for i := 0; i < las.nPoints; i++ { //loop by dept (.)
		fmt.Fprintf(&b, "%-9.3f ", dept.D[i])
		for j := 1; j < n; j++ { //loop by logs
			l, err = las.logByIndex(j)
			if err != nil {
				las.addWarning(TWarning{directOnWrite, lasSecData, i, "logByIndex() return error, log not found, panic"})
				return nil, errors.New("logByIndex() return error, log not found, panic")
			}
			fmt.Fprintf(&b, "%-9.3f ", l.V[i])
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
		return las.Null
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileStop(las.FileName, "~A")
	if (err != nil) || (iScanner == nil) {
		return las.Null
	}

	s := ""
	dept1 := 0.0
	for i := 0; iScanner.Scan(); i++ {
		s = strings.TrimSpace(iScanner.Text())
		if (len(s) == 0) || (s[0] == '#') {
			continue
		}
		k := strings.IndexRune(s, ' ')
		if k < 0 {
			k = len(s)
		}
		dept1, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			return las.Null
		}
		return dept1
	}
	//если мы попали сюда, то всё грусно, в файле после ~A не нашлось двух строчек с данными... или пустые строчки или комменты
	// TODO последняя строка "return las.Null" не обрабатывается в тесте
	return las.Null
}

// GetStepFromData - return step from data section
// read 2 line from section ~A and determine step
// close file
// return Null if error occure
func (las *Las) GetStepFromData() float64 {
	iFile, err := os.Open(las.FileName)
	if err != nil {
		return las.Null
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
		if (len(s) == 0) || (s[0] == '#') {
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
	//TODO последняя строка "return las.Null" не обрабатывается в тесте
	return las.Null
}

func (las *Las) setStep(h float64) {
	las.Step = h
}

func (las *Las) setStrt(h float64) {
	las.Strt = h
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
func (las *Las) SetNull(aNull float64) error {
	for _, l := range las.Logs { //loop by logs
		for i := range l.V { //loop by dept step
			if l.V[i] == las.Null {
				l.V[i] = aNull
			}
		}
	}
	las.Null = aNull
	return nil
}

//logByIndex - return log from map by Index
func (las *Las) logByIndex(i int) (*LasCurve, error) {
	for _, v := range las.Logs {
		if v.Index == i {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("log with index: %v not present", i)
}
