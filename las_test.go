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

//test special cases
//empty file name, file in wrong encoding
func TestLasOpenSpeсial(t *testing.T) {
	las := NewLas()
	n, err := las.Open("")

	assert.Equal(t, 0, n)
	assert.NotNil(t, err, fmt.Sprintf("<TestLasOpenSpeсial> expect error not nil, got '%v'\n", err))

	// If the file is not found different operating systems resport a different message
	// assert.Equal(t, "open : The system cannot find the file specified.", err.Error())
	errBool := false
	errMsg := err.Error()
	errMsgs := []string{"open : The system cannot find the file specified.", "open : no such file or directory"}
	for _, msg := range errMsgs {
		if msg == errMsg {
			errBool = true
		}
	}
	assert.True(t, errBool)

	n, err = las.Open(fp.Join("data/utf-32be-bom.las"))
	//this decode not support, return error
	assert.Equal(t, 0, n)
	assert.NotNil(t, err, fmt.Sprintf("<TestLasOpenSpeсial> expect error not nil, got '%v'\n", err))
	assert.Equal(t, "cpd: codepage not support encode/decode", err.Error())
}

// Проверка на достижение максимального количества варнингов
/* reading "more_20_warnings.las" generate warnings
 */
func TestReachingMaxAmountWarnings(t *testing.T) {
	las := NewLas()
	las.Open(fp.Join("data/more_20_warnings.las"))
	//by default maximum warning count is 20 + 1 about reaching
	assert.Equal(t, 21, las.Warnings.Count(), fmt.Sprintf("<TestReachingMaxAmountWarnings> on '%s' wrong count warning: %d\n", las.FileName, las.Warnings.Count()))

	saveMaxWarningCount := MaxWarningCount
	MaxWarningCount = 100
	las = NewLas()
	las.Open(fp.Join("data/more_20_warnings.las"))
	assert.Equal(t, 38, las.Warnings.Count(), fmt.Sprintf("<TestReachingMaxAmountWarnings> on file '%s' wrong warning number: %d expected 38\n", las.FileName, las.Warnings.Count()))
	MaxWarningCount = saveMaxWarningCount

	// SaveWarning() does not add to las.Warnings
	las.SaveWarning(fp.Join("data/more_20_warnings.wrn"))
	assert.Equal(t, 38, las.Warnings.Count(), fmt.Sprintf("<TestReachingMaxAmountWarnings> after las.SaveWarning() number warning changed: %d expected 38\n", las.Warnings.Count()))

	// test for error occure when SaveWarning() fails to write to the file
	assert.NotNil(t, las.SaveWarning(""))
}

// test SaveWarningToWriter
func TestLasSaveWarning(t *testing.T) {
	las := NewLas()
	las.Open(fp.Join("data/more_20_warnings.las"))
	f, err := os.Create(fp.Join("data/w1_more_20_warnings.txt"))
	buf := bufio.NewWriter(f)
	n := las.SaveWarningToWriter(buf)
	buf.Flush()
	f.Close()
	assert.Equal(t, 21, n)
	//now read 'w1_more_20_warnings.txt'
	//must be total 22 lines: 0-20 warnings, and last empty line with number 21
	f, err = os.Open(fp.Join("data/w1_more_20_warnings.txt"))
	assert.Nil(t, err)
	sc := bufio.NewScanner(f)
	i := 0
	for ; sc.Scan(); i++ {
	}
	f.Close()
	assert.Equal(t, 21, i)
}

type tGetDataStrt struct {
	fn string
	st float64
}

var dGetDataStrt = []tGetDataStrt{
	{fp.Join("data/2.0/sample_2.0_missing_strt.las"), 1670.000},
	{fp.Join("data/2.0/sample_2.0.las"), 1670.000},
	{fp.Join(""), -999.25},
}

func TestGetStrtFromData(t *testing.T) {
	for _, tmp := range dGetDataStrt {
		las := NewLas()
		las.Open(tmp.fn)
		assert.Equal(t, tmp.st, las.Strt, fmt.Sprintf("<TestGetStepFromData> fail on file '%s' \n", tmp.fn))
	}
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

func TestLasSetNull(t *testing.T) {
	las := NewLas()
	las.Open(fp.Join("data/expand_points_01.las"))
	assert.Equal(t, -9999.00, las.Null)
	las.SetNull(-999.25)
	assert.Equal(t, -999.25, las.Null)
	las.Save("-tmp.las")

	las = NewLas()
	las.Open("-tmp.las")
	assert.Equal(t, -999.25, las.Null)
	log := las.Logs[1]
	assert.Equal(t, las.Null, log.V[2])
	assert.Equal(t, las.Null, las.Logs[2].V[6])
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

		las = NewLas()
		n, err := las.Open(tmp.fn)
		//os.Remove(tmp.fn)
		assert.Nil(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, tmp.newNull, las.Null)
		assert.Equal(t, tmp.strt, las.Strt)
		assert.Equal(t, tmp.stop, las.Stop)
		assert.Equal(t, tmp.step, las.Step)
		assert.Equal(t, tmp.well, las.Well)
		assert.Equal(t, "DEPT", las.Logs[0].Name)
		assert.Equal(t, 1.1, las.Logs[0].D[1])
		assert.Equal(t, 1.0, las.Logs[0].V[0])
		assert.Equal(t, 1.2, las.Logs[1].D[2])
		assert.Equal(t, 2.2, las.Logs[1].V[2])
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
	assert.False(t, NewLas().IsEmpty())
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
	{cpd.CP1251, 0.0, 0.201, 10.01, 0.01, "Примерная-101 / бис", map[string]bool{"STOP": true, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": true, "WELL": true}},
	{cpd.CP1251, -99.99, 0.201, 10.01, 0.0, "Примерная-101 / бис", map[string]bool{"STOP": true, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": false, "NULL": true, "SSTP": true, "WELL": true}},
	{cpd.KOI8R, 0.0, 0.2, 2.0, 0.0, "Примерная-1001 /\"бис\"", map[string]bool{"STOP": true, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": false, "NULL": false, "SSTP": true, "WELL": true}},
	{cpd.CP866, 0.0, 0.21, 0.21, 0.1, "Примерная-101 /\"бис\"", map[string]bool{"STOP": true, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": false, "WELL": true}},
	{cpd.UTF8, 0.0, 0.2, 0.2, 0.0, "", map[string]bool{"STOP": true, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": false, "NULL": false, "SSTP": false, "WELL": false}},
	{cpd.UTF16LE, 0.0, 20.2, 1.0, -0.0, "", map[string]bool{"STOP": true, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": false, "NULL": false, "SSTP": true, "WELL": false}},
	{cpd.CP1251, 0.0, -999.25, 10.01, 0.01, "Примерная-101 / бис", map[string]bool{"STOP": true, "STPU": true, "STRT": false, "WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": true, "WELL": true}},
	{cpd.CP1251, 0.0, 0, -999.25, 0.01, "Примерная-101 / бис", map[string]bool{"STOP": false, "STPU": true, "STRT": true, "WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": true, "WELL": true}},
	{cpd.CP1251, 0.0, 0, 2.1, -999.25, "Примерная-101 / бис", map[string]bool{"STOP": true, "STPU": false, "STRT": true, "WRAP": true, "CURV": true, "STEP": true, "NULL": false, "SSTP": true, "WELL": true}},
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

//Тестирование заполнения секций
type tSecFill struct {
	fn string
	n  int //количество считанных точек данных
	well,
	null,
	ver,
	wrap string //версия las файла
}

var dSecFill = []tSecFill{
	{fp.Join("data/expand_points_01.las"), 7, "12-Сплошная", "-9999.00", "1.20", "NO"},
}

func TestSection(t *testing.T) {
	for _, tmp := range dSecFill {
		las := NewLas()
		n, err := las.Open(tmp.fn)
		assert.Nil(t, err, fmt.Sprintf("<TestExpandPoints> on '%s' return error: %v\n", tmp.fn, err))
		assert.Equal(t, tmp.n, n, fmt.Sprintf("<TestExpandPoints> on '%s' return n: %d expect: %d\n", tmp.fn, n, tmp.n))
		assert.Equal(t, tmp.n, las.NumPoints())
		assert.Equal(t, tmp.ver, las.VerSec.params["VERS"].Val)
		assert.Equal(t, tmp.wrap, las.VerSec.params["WRAP"].Val)
		assert.Equal(t, tmp.null, las.WelSec.params["NULL"].Val)
		assert.Equal(t, tmp.well, las.WelSec.params["WELL"].Val)
	}
}
