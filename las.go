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

// LasWellInfo - contain parameters of Well section
type LasWellInfo struct {
	ver       float64
	wrap      string
	wellName  string
	null      float64
	oCodepage cpd.IDCodePage
}

// Las - class to store las file
// input code page autodetect
// at read file always code page converted to UTF
// at save file code page converted to specifyed in Las.toCodePage
//TODO add pointer to cfg
//TODO при создании объекта las есть возможность указать кодировку записи, нужна возможность указать явно кодировку чтения
type Las struct {
	FileName        string             //file name from load
	File            *os.File           //the file from which we are reading
	Reader          io.Reader          //reader created from File, provides decode from codepage to UTF-8
	scanner         *bufio.Scanner     //scanner
	Ver             float64            //version 1.0, 1.2 or 2.0
	Wrap            string             //YES || NO
	Strt            float64            //start depth
	Stop            float64            //stop depth
	Step            float64            //depth step
	Null            float64            //value interpreted as empty
	Well            string             //well name
	Rkb             float64            //altitude KB
	Logs            LasCurves          //store all logs
	LogDic          *map[string]string //external dictionary of standart log name - mnemonics
	VocDic          *map[string]string //external vocabulary dictionary of log mnemonic
	Warnings        TLasWarnings       //slice of warnings occure on read or write
	ePoints         int                //expected count (.)
	nPoints         int                //actually count (.)
	oCodepage       cpd.IDCodePage     //codepage to save, default xlib.CpWindows1251. to special value, specify at make: NewLas(cp...)
	iDuplicate      int                //индекс повторящейся мнемоники, увеличивается на 1 при нахождении дубля, начально 0
	currentLine     int                //index of current line in readed file
	maxWarningCount int                //default maximum warning count
	stdNull         float64            //default null value
}

// GetStepFromData - return step from data section
// read 2 line from section ~A and determine step
// close file
// return Null if error occure
// если делать функцией, не методом, то придётся NULL передавать. а оно надо вообще
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
			return las.Null
		}
		j++
		if j == 2 {
			return math.Round((dept1-dept2)*10) / 10
		}
		dept2 = dept1
	}
	//если мы попали сюда, то всё грусно, в файле после ~A не нашлось двух строчек с данными... или пустые строчки или комменты
	// TODO последняя строка "return las.Null" не обрабатывается в тесте
	return las.Null
}

func (las *Las) setStep(h float64) {
	las.Step = h
}

//SetNull - change parameter NULL in WELL INFO section and in all logs
func (las *Las) SetNull(aNull float64) error {
	for _, l := range las.Logs { //loop by logs
		for i := range l.log { //loop by dept step
			if l.log[i] == las.Null {
				l.log[i] = aNull
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

var (
	// ExpPoints - ожидаемое количество точек данных, до чтения мы не можем знать сколько точек будет фактически прочитано
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
	las.ePoints = 1000
	las.Logs = make(map[string]LasCurve)
	las.maxWarningCount = MaxWarningCount
	las.stdNull = -999.25
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

/*
// NewLasPar - create new object with parameters
func NewLasPar(lasInfo LasWellInfo) *Las {
	las := new(Las)
	las.Ver = lasInfo.ver
	las.Wrap = lasInfo.wrap
	las.Well = lasInfo.wellName
	las.stdNull = lasInfo.null
	las.oCodepage = lasInfo.oCodepage

	las.maxWarningCount = MaxWarningCount
	las.ePoints = ExpPoints
	las.Logs = make(map[string]LasCurve)
	//mnemonic dictionary
	las.LogDic = nil
	//external log dictionary
	las.VocDic = nil
	//счётчик повторяющихся мнемоник, увеличивается каждый раз на 1, используется при переименовании мнемоники
	las.iDuplicate = 0
	return las
}
*/

// selectSection - analize first char after ~
// ~V - section vertion
// ~W - well info section
// ~C - curve info section
// ~A - data section
func (las *Las) selectSection(r rune) int {
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

const (
	checkHeaderStep     = iota
	checkHeaderNull     = iota
	checkHeaderWrap     = iota
	checkHeaderCurve    = iota
	checkHeaderStrtStop = iota
)

// HeaderCheckMsg - one message on check las Header
type HeaderCheckMsg struct {
	id  int
	msg string
}

// HeaderCheckRes - result of check readed las Header
type HeaderCheckRes []HeaderCheckMsg

func (hc *HeaderCheckRes) needUpdateStep() bool {
	for _, m := range *hc {
		if m.id == checkHeaderStep {
			return true
		}
	}
	return false
}

func (hc *HeaderCheckRes) needUpdateNull() bool {
	for _, m := range *hc {
		if m.id == checkHeaderNull {
			return true
		}
	}
	return false
}

func (hc *HeaderCheckRes) addStepWarning() {
	*hc = append(*hc, HeaderCheckMsg{checkHeaderStep, ""})
}

func (hc *HeaderCheckRes) addNullWarning() {
	*hc = append(*hc, HeaderCheckMsg{checkHeaderNull, ""})
}

// make test of loaded las header
// return error:
// - double error on STEP parameter
// - las file is WRAP == ON
// - Curve section not exist
func (las *Las) checkHeader() (HeaderCheckRes, error) {
	res := make(HeaderCheckRes, 0)
	if las.Null == 0.0 {
		res.addNullWarning()
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("NULL parameter equal 0, replace to %4.3f", las.Null)})
	}
	if las.Step == 0.0 {
		res.addStepWarning()
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("STEP parameter equal 0, replace to %4.3f", las.Step)})
	}
	if math.Abs(las.Stop-las.Strt) < 0.1 {
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid STRT: %4.3f or STOP: %4.3f, will be replace to actually", las.Strt, las.Stop)})
	}
	if las.IsWraped() {
		las.addWarning(TWarning{directOnRead, lasSecData, -1, "WRAP = YES, file ignored"})
		return res, fmt.Errorf("Wrapped files not support") //return 0, nil
	}
	if len(las.Logs) <= 0 {
		las.addWarning(TWarning{directOnRead, lasSecData, -1, "section ~Curve not exist, file ignored"})
		return res, fmt.Errorf("Curve section not exist") //return 0, nil
	}
	return res, nil
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
// - file to open not exist
// - file cannot be decoded to UTF-8
// - las is wrapped
// - las file not contain Curve section
func (las *Las) Open(fileName string) (int, error) {
	var err error
	las.File, err = os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer las.File.Close()
	las.FileName = fileName
	//create and store Reader, this reader decode to UTF-8
	las.Reader, err = cpd.NewReader(las.File)
	if err != nil {
		return 0, err
	}
	las.scanner = bufio.NewScanner(las.Reader)
	las.currentLine = 0
	las.LoadHeader()
	// проверки на фатальные ошибки
	wrongChecker := NewWrongChecker()
	r := wrongChecker.check(las)
	if r.wrapWrong() {
		las.addWarning(TWarning{directOnRead, lasSecData, -1, "WRAP = YES, file ignored"})
		return 0, fmt.Errorf("Wrapped files not support")
	}
	if r.curvesWrong() {
		las.addWarning(TWarning{directOnRead, lasSecData, -1, "section ~Curve not exist, file ignored"})
		return 0, fmt.Errorf("Curve section not exist")
	}
	// фатальные ошибки в заголовке исключены
	// делаем стандартные проверки заголовка
	stdChecker := NewStdChecker()
	r = stdChecker.check(las)
	if r.nullWrong() {
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("NULL parameter equal 0, replace to %4.3f", las.Null)})
		las.SetNull(las.stdNull)
	}
	if r.strtStopWrong() {
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid STRT: %4.3f == STOP: %4.3f, will be replace to actually", las.Strt, las.Stop)})
	}

	if r.stepWrong() {
		las.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("STEP parameter equal 0, replace to %4.3f", las.Step)})
		h := las.GetStepFromData() // return o.Null if cannot calculate step from data
		if h == las.Null {
			return 0, errors.New("invalid STEP parameter and invalid step in data")
		}
		las.setStep(h)
	}
	return las.ReadDataSec(fileName)
}

/*
func (o *Las) open1(fileName string) (int, error) {
	var err error
	o.File, err = os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer o.File.Close()
	o.FileName = fileName
	//create and store Reader, this reader decode to UTF-8
	o.Reader, err = cpd.NewReader(o.File)
	if err != nil {
		return 0, err
	}
	o.scanner = bufio.NewScanner(o.Reader)
	o.currentLine = 0
	o.LoadHeader()

	// проверка корректности данных секции WELL INFO перез загрузкой данных
	res, err := o.checkHeader() // res содержит несколько сообщений связанных с корректностью заголовка las файла
	if err != nil {
		return 0, err // дальше читать файл смысла нет, или файл с переносами или нет секции Curve ...
	}
	// обрабатываем изменение параметров las файла по результатам чтения заголовка
	if res.needUpdateNull() {
		o.SetNull(o.stdNull)
	}
	if res.needUpdateStep() {
		h := o.GetStepFromData() // return o.Null if cannot calculate step from data
		o.setStep(h)
		if h == o.Null {
			return 0, errors.New("invalid STEP parameter and invalid step in data")
		}
	}
	return o.ReadDataSec(fileName)
}
*/

/*LoadHeader - read las file and load all section before ~A
   secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - A data
1. читаем строку
2. если коммент или пустая в игнор
3. если начало секции, определяем какой
4. если началась секция данных заканчиваем
5. читаем одну строку (это один параметер из известной нам секции)
   Пока ошибку всегда возвращает nil, причин возвращать другое значение пока нет.
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
				break // dAta section read after //exit from for
			}
		} else {
			err = las.ReadParameter(s, secNum) //if not comment, not empty and not new section => parameter, read it
			if err != nil {
				las.addWarning(TWarning{directOnRead, secNum, -1, fmt.Sprintf("while process parameter: '%s' occure error: %v", s, err)})
			}
		}
	}
	return nil
}

// ReadParameter - read one parameter
func (las *Las) ReadParameter(s string, secNum int) error {
	switch secNum {
	case lasSecVertion:
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
			//las.Well = p.Val
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
	l.Init(len(las.Logs), las.GetMnemonic(l.Name), las.ChangeDuplicateLogName(l.Name), las.GetExpectedPointsCount())
	las.Logs[l.Name] = l //добавление в хранилище кривой каротажа с колонкой глубин
	return nil
}

//GetExpectedPointsCount - оценка количества точек по параметрам STEP, STRT, STOP
func (las *Las) GetExpectedPointsCount() int {
	var m int
	//TODO нужно обработать все случаи
	if las.Step == 0.0 {
		return las.ePoints
	}
	if math.Abs(las.Stop) > math.Abs(las.Strt) {
		m = int((las.Stop-las.Strt)/las.Step) + 2
	} else {
		m = int((las.Strt-las.Stop)/las.Step) + 2
	}
	if m < 0 {
		m = -m
	}
	if m == 0 {
		return las.ePoints
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
	copy(newDept, d.dept)
	d.dept = newDept

	newLog := make([]float64, las.ePoints)
	copy(newLog, d.dept)
	d.log = newLog
	las.Logs[d.Name] = *d

	//loop over other logs
	n := len(las.Logs)
	var l *LasCurve
	for j := 1; j < n; j++ {
		l, _ = las.logByIndex(j)
		newDept := make([]float64, las.ePoints)
		copy(newDept, l.dept)
		l.dept = newDept

		newLog := make([]float64, las.ePoints)
		copy(newLog, l.log)
		l.log = newLog
		las.Logs[l.Name] = *l
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

	//исходя из параметров STRT, STOP и STEP определяем ожидаемое количество строк данных
	las.ePoints = las.GetExpectedPointsCount()
	//o.currentLine++
	n := len(las.Logs)       //количество каротажей, столько колонок данных ожидаем
	d, _ = las.logByIndex(0) //dept log
	s := ""
	for i = 0; las.scanner.Scan(); i++ {
		las.currentLine++
		if i == las.ePoints {
			las.expandDept(d)
		}
		s = strings.TrimSpace(las.scanner.Text())
		// i счётчик не строк, а фактически считанных данных - счётчик добавлений в слайсы данных
		//TODO возможно следует завести отдельный счётчик и оставить в покое счётчик цикла
		if isIgnoredLine(s) {
			i--
			continue
		}
		//first column is DEPT
		k := strings.IndexRune(s, ' ') //TODO вероятно получим ошибку если данные будут разделены не пробелом а табуляцией или ещё чем-то
		if k < 0 {                     //line must have n+1 column and n separated spaces block (+1 becouse first column DEPT)
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
		d.dept[i] = dept
		// проверка шага у первых двух точек данных и сравнение с параметром step
		//TODO данную проверку следует делать через Checker
		if i > 1 {
			if math.Pow(((dept-d.dept[i-1])-las.Step), 2) > 0.1 {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("actual step %5.2f ≠ global STEP %5.2f", (dept - d.dept[i-1]), las.Step)})
			}
		}
		// проверка шага между точками [i-1, i] и точками [i-2, i-1] обнаружение немонотонности колонки глубин
		if i > 2 {
			if math.Pow(((dept-d.dept[i-1])-(d.dept[i-1]-d.dept[i-2])), 2) > 0.1 {
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, fmt.Sprintf("step %5.2f ≠ previously step %5.2f", (dept - d.dept[i-1]), (d.dept[i-1] - d.dept[i-2]))})
				dept = d.dept[i-1] + las.Step
			}
		}

		s = strings.TrimSpace(s[k+1:]) //cut first column
		//цикл по каротажам
		for j := 1; j < (n - 1); j++ {
			iSpace := strings.IndexRune(s, ' ')
			switch iSpace {
			case -1: //не все колонки прочитаны, а пробелов уже нет... пробуем игнорировать сроку заполняя оставшиеся каротажи NULLами
				las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "not all column readed, set log value to NULL"})
			case 0:
				v = las.Null
			case 1:
				v, err = strconv.ParseFloat(s[:1], 64)
			default:
				v, err = strconv.ParseFloat(s[:iSpace], 64) //strconv.ParseFloat(s[:iSpace-1], 64)
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
			l.dept[i] = dept
			l.log[i] = v
			s = strings.TrimSpace(s[iSpace+1:])
		}
		//остаток - последняя колонка
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			las.addWarning(TWarning{directOnRead, lasSecData, las.currentLine, "not all column readed, set log value to NULL"})
			v = las.Null
		}
		l, err = las.logByIndex(n - 1)
		if err != nil {
			las.nPoints = i
			return i, errors.New("internal ERROR, func (las *Las) readDataSec()::las.logByIndex(j) return error on last column")
		}
		l.dept[i] = dept
		l.log[i] = v
	}
	//i - actually readed lines and add (.) to data array
	//crop logs to actually len
	//TODO перенести в Open()
	err = las.setActuallyNumberPoints(i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// NumPoints - return actually number of points in data
func (las *Las) NumPoints() int {
	return las.nPoints
}

//Dept - return slice of DEPT curve (first column)
func (las *Las) Dept() []float64 {
	d, err := las.logByIndex(0)
	if err != nil {
		return nil
	}
	return d.dept
}

func (las *Las) setActuallyNumberPoints(numPoints int) error {
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

/*
func (o *Las) convertStrToOut(s string) string {
	r, _ := cpd.NewReaderTo(strings.NewReader(s), o.oCodepage.String())
	b, _ := ioutil.ReadAll(r)
	return string(b)
}


func (o *Las) SaveToFile(fileName string, useMnemonic ...bool) error {
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
*/

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
		bufToSave, err = las.SaveToBuf(true)
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
	fmt.Fprintf(&b, _LasVersion, las.Ver)
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
		fmt.Fprintf(&b, "%-9.3f ", dept.dept[i])
		for j := 1; j < n; j++ { //loop by logs
			l, err = las.logByIndex(j)
			if err != nil {
				las.addWarning(TWarning{directOnWrite, lasSecData, i, "logByIndex() return error, log not found, panic"})
				return nil, errors.New("logByIndex() return error, log not found, panic")
			}
			fmt.Fprintf(&b, "%-9.3f ", l.log[i])
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
