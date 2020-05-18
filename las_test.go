//(c) softland 2020
//softlandia@gmail.com
package glasio

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	fp "path/filepath"

	"github.com/softlandia/cpd"
	"github.com/stretchr/testify/assert"
)

func TestGetMnemonic(t *testing.T) {
	Mnemonic, err := LoadStdMnemonicDic(fp.Join("data/mnemonic.ini"))
	assert.Nil(t, err, fmt.Sprintf("load std mnemonic error: %v\n check out 'data\\mnemonic.ini'", err))

	VocDic, err := LoadStdVocabularyDictionary(fp.Join("data/dic.ini"))
	assert.Nil(t, err, fmt.Sprintf("load std vocabulary dictionary error: %v\n check out 'data\\dic.ini'", err))

	las := NewLas()
	mnemonic := las.GetMnemonic("1")
	assert.Equal(t, mnemonic, "", fmt.Sprintf("<GetMnemonic> return '%s', expected ''\n", mnemonic))

	las.LogDic = &Mnemonic
	las.VocDic = &VocDic

	mnemonic = las.GetMnemonic("GR")
	assert.Equal(t, mnemonic, "GR", fmt.Sprintf("<GetMnemonic> return '%s', expected 'GR'\n", mnemonic))

	mnemonic = las.GetMnemonic("ГК")
	assert.Equal(t, mnemonic, "GR", fmt.Sprintf("<GetMnemonic> return '%s', expected 'GR'\n", mnemonic))
}

//Проверка на достижение максимального количества варнингов
//По умолчанию MaxNumWarinig = 20
func TestReachingMaxAmountWarnings(t *testing.T) {
	las := NewLas()
	las.Open(fp.Join("data/more_20_warnings.las"))
	assert.GreaterOrEqual(t, las.Warnings.Count(), 20, fmt.Sprintf("<TestReachingMaxAmountWarnings> on read file data\\more_20_warnings.las warning count: %d\n", las.Warnings.Count()))

	ExpPoints = 2
	las = NewLas()
	las.maxWarningCount = 100
	las.Open(fp.Join("data/more_20_warnings.las"))
	if las.Warnings.Count() != 41 {
		las.SaveWarning(fp.Join("data/more_20_warnings.wrn"))
		assert.Equal(t, 41, las.Warnings.Count(), fmt.Sprintf("<TestReachingMaxAmountWarnings> on read file data\\more_20_warnings.las warning count: %d expected 62\n", las.Warnings.Count()))
	}

	assert.NotNil(t, las.SaveWarning("<wrn>.md"))
}

//тестируем особые случаи открытия
//пустые имена файлов, файлы в неправильной кодировке
func TestLasOpenSpeсial(t *testing.T) {
	las := NewLas()
	n, err := las.Open("")
	assert.Equal(t, 0, n)
	assert.NotNil(t, err, fmt.Sprintf("<TestLasOpenSpeсial> expect error not nil, got '%v'\n", err))
	assert.Equal(t, "open : The system cannot find the file specified.", err.Error())

	n, err = las.Open(fp.Join("data/utf-32be-bom.las"))
	//this decode not support, return error
	assert.Equal(t, 0, n)
	assert.NotNil(t, err, fmt.Sprintf("<TestLasOpenSpeсial> expect error not nil, got '%v'\n", err))
	assert.Equal(t, "cpd: codepage not support encode/decode", err.Error())
}

type tLoadHeader struct {
	fn   string
	ver  float64
	wrap string
	strt float64
	stop float64
	step float64
	null float64
	well string
}

var dLoadHeader = []tLoadHeader{
	{fp.Join("data/2.0/cp1251_2.0_well_name.las"), 2.0, "NO", 0.0, 39.9, 0.3, -999.25, "Примерная-1 / бис(ё)"},
	{fp.Join("data/2.0/cp1251_2.0_based.las"), 2.0, "NO", 0.0, 39.9, 0.3, -999.25, "Примерная-1/бис(ё)"},
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1.0, 1.0, 0.1, -9999.00, "12-Сплошная"},
	{fp.Join("data/more_20_warnings.las"), 1.2, "NO", 0.0, 0.0, 0.0, -32768.0, "6"}, //in las file STEP=0.0 but this incorrect, LoadHeader replace STEP to actual from data
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1.0, 1.0, 0.1, -9999.0, "12-Сплошная"},
	{fp.Join("data/1.2/sample.las"), 1.2, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "ANY ET AL OIL WELL #12"},
	{fp.Join("data/2.0/sample_2.0.las"), 2.0, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "AAAAA_2"},
	{fp.Join("data/duplicate_step.las"), 1.2, "NO", 1670.0, 1660.0, -0.1200, -999.2500, "ANY ET AL OIL WELL #12"}, //duplicate_step.las contains two line with STEP:: STEP.M -0.1250: STEP.M -0.1200: using LAST parameter
	{fp.Join("data/encodings_utf8.las"), 1.2, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "Скв #12Ω"},
}

func TestLoadHeader(t *testing.T) {
	var las *Las
	for _, tmp := range dLoadHeader {
		las = NewLas()
		f, _ := os.Open(tmp.fn)
		las.Reader, _ = cpd.NewReader(f)
		las.FileName = tmp.fn
		las.scanner = bufio.NewScanner(las.Reader)
		las.LoadHeader()
		assert.Equal(t, tmp.ver, las.Ver, fmt.Sprintf("<LoadHeader> file '%s' readed VER: %f, expected %f", las.FileName, las.Ver, tmp.ver))
		assert.Equal(t, tmp.wrap, las.Wrap, fmt.Sprintf("<LoadHeader> file '%s' readed WRAP: %s, expected %s", las.FileName, las.Wrap, tmp.wrap))
		assert.Equal(t, tmp.strt, las.Strt, fmt.Sprintf("<LoadHeader> file '%s' readed STRT: %f, expected %f", las.FileName, las.Strt, tmp.strt))
		assert.Equal(t, tmp.stop, las.Stop, fmt.Sprintf("<LoadHeader> file '%s' readed STOP: %f, expected %f", las.FileName, las.Stop, tmp.stop))
		assert.Equal(t, tmp.step, las.Step, fmt.Sprintf("<LoadHeader> file '%s' readed STEP: %f, expected %f", las.FileName, las.Step, tmp.step))
		assert.Equal(t, tmp.null, las.Null, fmt.Sprintf("<LoadHeader> file '%s' readed NULL: %f, expected %f", las.FileName, las.Null, tmp.null))
		assert.Equal(t, tmp.well, las.Well, fmt.Sprintf("<LoadHeader> file '%s' readed WELL: %s, expected %s", las.FileName, las.Well, tmp.well))
	}
}

func TestLoadLasHeader(t *testing.T) {
	for _, tmp := range dLoadHeader {
		las, err := LoadLasHeader(tmp.fn)
		assert.Nil(t, err)
		assert.Equal(t, tmp.ver, las.Ver, fmt.Sprintf("<LoadHeader> file '%s' readed VER: %f, expected %f", las.FileName, las.Ver, tmp.ver))
		assert.Equal(t, tmp.wrap, las.Wrap, fmt.Sprintf("<LoadHeader> file '%s' readed WRAP: %s, expected %s", las.FileName, las.Wrap, tmp.wrap))
		assert.Equal(t, tmp.strt, las.Strt, fmt.Sprintf("<LoadHeader> file '%s' readed STRT: %f, expected %f", las.FileName, las.Strt, tmp.strt))
		assert.Equal(t, tmp.stop, las.Stop, fmt.Sprintf("<LoadHeader> file '%s' readed STOP: %f, expected %f", las.FileName, las.Stop, tmp.stop))
		assert.Equal(t, tmp.step, las.Step, fmt.Sprintf("<LoadHeader> file '%s' readed STEP: %f, expected %f", las.FileName, las.Step, tmp.step))
		assert.Equal(t, tmp.null, las.Null, fmt.Sprintf("<LoadHeader> file '%s' readed NULL: %f, expected %f", las.FileName, las.Null, tmp.null))
		assert.Equal(t, tmp.well, las.Well, fmt.Sprintf("<LoadHeader> file '%s' readed WELL: %s, expected %s", las.FileName, las.Well, tmp.well))
	}
	//test error case
	las, err := LoadLasHeader("not_exist_file.las") //file not exist
	assert.NotNil(t, err)
	assert.Nil(t, las)
	las, err = LoadLasHeader(fp.Join("data/utf-32be-bom.las")) //file exist, codepage not support
	assert.NotNil(t, err)
	assert.Nil(t, las)
}

func TestLasSaveWarning(t *testing.T) {
	las := NewLas()
	las.Open(fp.Join("data/more_20_warnings.las"))
	err := las.SaveWarning(fp.Join("data/w1_more_20_warnings.txt"))
	assert.Nil(t, err)
	f, _ := os.Create("nul")
	buf := bufio.NewWriter(f)
	n := las.SaveWarningToWriter(buf)
	f.Close()
	assert.Equal(t, 21, n)
}

type tGetDataStep struct {
	fn string
	st float64
}

var dGetDataStep = []tGetDataStep{
	{fp.Join("data/step-2-data-without-step-case1.las"), -32768},
	{fp.Join("data/step-2-data-without-step-case2.las"), -32768.000},
	{fp.Join("data/no-data-section.las"), -32768.000},
	{fp.Join("data/step-1-normal-case.las"), 1.0},
}

func TestGetStepFromData(t *testing.T) {
	for _, tmp := range dGetDataStep {
		las := NewLas()
		las.Open(tmp.fn)
		assert.Equal(t, tmp.st, las.Step, fmt.Sprintf("<TestGetStepFromData> fail on file '%s' \n", tmp.fn))
	}
}

//Тестирование увеличения чоличества точек
type tExpandDept struct {
	fn   string
	n    int //количество считанных точек данных
	nWrn int //количество предупреждений
}

var dExpandDept = []tExpandDept{
	{fp.Join("data/expand_points_01.las"), 7, 5},
}

func TestExpandPoints(t *testing.T) {
	// перед тестом установим маленькое количество ожидаемых точек, иначе надо делать огномный тестовый файл
	// срабатывание добавления выполняется при переполнении буфера 1000
	ExpPoints = 2
	for _, tmp := range dExpandDept {
		las := NewLas()
		n, err := las.Open(tmp.fn)
		assert.Nil(t, err, fmt.Sprintf("<TestExpandPoints> on '%s' return error: %v\n", tmp.fn, err))
		assert.Equal(t, tmp.n, n, fmt.Sprintf("<TestExpandPoints> on '%s' return n: %d expect: %d\n", tmp.fn, n, tmp.n))
		assert.Equal(t, tmp.n, las.NumPoints())
		assert.Equal(t, tmp.nWrn, las.Warnings.Count(), fmt.Sprintf("<TestExpandPoints> '%s' return warning count %d, expected %d\n", tmp.fn, las.Warnings.Count(), tmp.nWrn))
		assert.Contains(t, las.Warnings[2].String(), "line: 25", fmt.Sprintf("<TestExpandPoints> '%s' return: '%s' wrong warning index 2\n", tmp.fn, las.Warnings[2]))
		assert.Contains(t, las.Warnings[4].String(), "line: 27", fmt.Sprintf("<TestExpandPoints> '%s' return: '%s' wrong warning index 4\n", tmp.fn, las.Warnings[4]))
	}
}

func TestLasSetNull(t *testing.T) {
	las := NewLas()
	las.Open(fp.Join("data/expand_points_01.las"))
	assert.Equal(t, -9999.00, las.Null)
	las.SetNull(-999.25)
	assert.Equal(t, -999.25, las.Null)
	las.Save("-tmp.las")
	las.Open("-tmp.las")
	assert.Equal(t, -999.25, las.Null)
	log := las.Logs["аПС"]
	assert.Equal(t, las.Null, log.log[2])
	assert.Equal(t, las.Null, las.Logs["аПС2"].log[6])
	err := os.Remove("-tmp.las")
	assert.Nil(t, err, fmt.Sprintf("%v", err))
}

type tSaveLas struct {
	fn      string
	cp      cpd.IDCodePage
	null    float64
	strt    float64
	stop    float64
	step    float64
	well    string
	newNull float64
}

var dSaveLas = []tSaveLas{
	//        filename                codepage    null    strt   stop  step   well name           new null
	{fp.Join("test_files/~1251.las"), cpd.CP1251, -99.99, 0.201, 10.01, 0.01, "Примерная-101 / бис", -0.1},
	{fp.Join("test_files/~koi8.las"), cpd.KOI8R, -99.0, 0.2, 2.0, -0.1, "Примерная-1001 /\"бис\"", -55.55},
	{fp.Join("test_files/~866.las"), cpd.CP866, -909.0, 2.21, 12.1, -0.1, "Примерная-101 /\"бис\"", 5555.55},
	{fp.Join("test_files/~utf-8.las"), cpd.UTF8, -999.99, 20.21, 1.0, -0.01, "Примерная-101А / бис", -999.25},
	{fp.Join("test_files/~utf-16le.las"), cpd.UTF16LE, -999.99, 20.2, 1.0, -0.01, "Примерная-101А /бис", -999.25},
}

// проверяется запись
// дополнительно проверяем функцию SetNull, это позволяет изменить содержимое las
func TestLasSave(t *testing.T) {
	for _, tmp := range dSaveLas {
		las := makeSampleLas(tmp.cp, tmp.null, tmp.strt, tmp.stop, tmp.step, tmp.well)
		las.SetNull(tmp.newNull)
		err := las.Save(tmp.fn)
		assert.Nil(t, err)

		n, err := las.Open(tmp.fn)
		//os.Remove(tmp.fn)
		assert.Nil(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, tmp.newNull, las.Null)
		assert.Equal(t, tmp.strt, las.Strt)
		assert.Equal(t, tmp.stop, las.Stop)
		assert.Equal(t, tmp.step, las.Step)
		assert.Equal(t, tmp.well, las.Well)
	}
}

func TestSetNullOnEmptyLas(t *testing.T) {
	las := NewLas()
	las.SetNull(-1000)
	assert.Equal(t, -1000.0, las.Null)
}

func TestLasIsEmpty(t *testing.T) {
	las := Las{}
	assert.True(t, las.IsEmpty())
	las = *NewLas()
	assert.False(t, las.IsEmpty())
}

type tStdCheckLas struct {
	cp       cpd.IDCodePage
	null     float64
	strt     float64
	stop     float64
	step     float64
	well     string
	testsRes map[string]bool // stdCheck contain 4 test
}

var dCheckLas = []tStdCheckLas{
	//codepage   null strt   stop  step   well name             проверки  wrap, curve, step  null  strt   well
	{cpd.CP1251, 0.0, 0.201, 10.01, 0.01, "Примерная-101 / бис", map[string]bool{"WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": true, "WELL": true}},
	{cpd.CP1251, -99.99, 0.201, 10.01, 0.0, "Примерная-101 / бис", map[string]bool{"WRAP": true, "CURV": true, "STEP": false, "NULL": true, "SSTP": true, "WELL": true}},
	{cpd.KOI8R, 0.0, 0.2, 2.0, 0.0, "Примерная-1001 /\"бис\"", map[string]bool{"WRAP": true, "CURV": true, "STEP": false, "NULL": false, "SSTP": true, "WELL": true}},
	{cpd.CP866, 0.0, 0.21, 0.21, 0.1, "Примерная-101 /\"бис\"", map[string]bool{"WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": false, "WELL": true}},
	{cpd.UTF8, 0.0, 0.2, 0.2, 0.0, "", map[string]bool{"WRAP": true, "CURV": true, "STEP": false, "NULL": false, "SSTP": false, "WELL": false}},
	{cpd.UTF16LE, 0.0, 20.2, 1.0, -0.0, "", map[string]bool{"WRAP": true, "CURV": true, "STEP": false, "NULL": false, "SSTP": true, "WELL": false}},
}

func TestLasChecker(t *testing.T) {
	chkr := NewStdChecker() //стандартная проверка на step=0, null=0, strt=stop, well=""
	for i, tmp := range dCheckLas {
		las := makeSampleLas(tmp.cp, tmp.null, tmp.strt, tmp.stop, tmp.step, tmp.well)
		assert.NotEqual(t, 0, len(chkr))
		for key, chk := range chkr {
			checkRes := chk.do(chk, las)
			assert.Equal(t, tmp.testsRes[key], checkRes.res, fmt.Sprintf("i:%d, r:%s", i, key))
		}
	}
}

func TestLasChecker2(t *testing.T) {
	stdChecker := NewStdChecker()

	tmp := dCheckLas[0] // в данных одна ошиба, NULL=0
	las := makeSampleLas(tmp.cp, tmp.null, tmp.strt, tmp.stop, tmp.step, tmp.well)
	res := stdChecker.check(las)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, res["NULL"].name, "NULL")
	assert.True(t, res.nullWrong(), fmt.Sprintf("%v", res.nullWrong()))
	assert.Contains(t, res["NULL"].warning.String(), "__WRN__ NULL parameter equal 0")
	assert.False(t, res.stepWrong())

	tmp = dCheckLas[4] // 4 ошибки NULL=0, START=STOP, STEP=0, WELL=''
	las = makeSampleLas(tmp.cp, tmp.null, tmp.strt, tmp.stop, tmp.step, tmp.well)
	res = stdChecker.check(las)
	assert.Equal(t, 4, len(res))
	assert.Equal(t, res["STEP"].name, "STEP")
	assert.Contains(t, res["WELL"].String(), "name: WELL,")
	assert.True(t, res.stepWrong())
	assert.True(t, res.nullWrong())
	assert.True(t, res.stepWrong())
	assert.True(t, res.wellWrong())
	assert.True(t, res.strtStopWrong(), fmt.Sprintf("%v", res["SSTP"]))
	assert.False(t, res.curvesWrong(), fmt.Sprintf("%v", res["CURV"]))
	assert.False(t, res.wrapWrong(), fmt.Sprintf("%v", res["WRAP"]))

	las = makeSampleLas(cpd.CP866, -999.25, 0, 100, 0.2, "well") //правильные данные
	res = stdChecker.check(las)                                  //StdChecker должен вернуть пустой слайс
	assert.Equal(t, 0, len(res))
}

func BenchmarkSave1(b *testing.B) {
	for _, tmp := range dSaveLas {
		las := makeSampleLas(tmp.cp, tmp.null, tmp.strt, tmp.stop, tmp.step, tmp.well)
		las.SetNull(tmp.newNull)
		las.Save(tmp.fn)
	}
}

/*
func BenchmarkSave2(b *testing.B) {
	for _, tmp := range dSaveLas {
		las := makeSampleLas(tmp.cp, tmp.null, tmp.strt, tmp.stop, tmp.step, tmp.well)
		las.SetNull(tmp.newNull)
		las.SaveToFile(tmp.fn)
	}
}
*/
