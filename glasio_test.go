//(c) softland 2019-2020
//softlandia@gmail.com

package glasio

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"testing"

	"github.com/softlandia/cpd"

	"github.com/stretchr/testify/assert"
)

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
	{fp.Join("data/step-2-data-without-step-case1.las"), -32768.000},
	{fp.Join("data/step-2-data-without-step-case2.las"), -32768.000},
	{fp.Join("data/no-data-section.las"), -32768.000},
	{fp.Join("data/step-1-normal-case.las"), 1.0},
}

func TestGetStepFromData(t *testing.T) {
	for _, tmp := range dGetDataStep {
		las := NewLas()
		las.Open(tmp.fn)
		assert.Equal(t, tmp.st, las.Step)
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
	for _, tmp := range dExpandDept {
		las := NewLas()
		n, err := las.Open(tmp.fn)
		assert.Nil(t, err, fmt.Sprintf("<TestExpandPoints> on '%s' return error: %v\n", tmp.fn, err))
		assert.Equal(t, n, tmp.n, fmt.Sprintf("<TestExpandPoints> on '%s' return n: %d expect: %d\n", tmp.fn, n, tmp.n))
		assert.Equal(t, las.Warnings.Count(), tmp.nWrn, fmt.Sprintf("<TestExpandPoints> '%s' return warning count %d, expected %d\n", tmp.fn, las.Warnings.Count(), tmp.nWrn))
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
	cp      cpd.IDCodePage
	null    float64
	strt    float64
	stop    float64
	step    float64
	well    string
	newNull float64
}

var dSaveLas = []tSaveLas{
	{cpd.CP1251, -99.99, 0.201, 10.01, 0.01, "Примерная-101/бис", -0.1},
	{cpd.CP866, -909.0, 2.21, 11.11, -0.1, "Примерная-101/\"бис\"", 5555.55},
	//{cpd.KOI8R, -99.0, 0.2, 2.0, -0.1, "Примерная-101/\"бис\"", -55.55},
	{cpd.UTF8, -999.99, 20.21, 1.0, -0.01, "Примерная-101А/бис", -999.25},
}

// проверяется запись
// первый раз в кодировке 1251
// второй раз в кодировке 866
// третий в UTF-8
// четвертый в KOI8
// дополнительно проверяем функцию SetNull
func TestLasSave(t *testing.T) {
	var las *Las
	for _, tmp := range dSaveLas {
		if tmp.cp == cpd.CP1251 {
			las = NewLas()
		} else {
			las = NewLas(tmp.cp)
		}
		las.Null = tmp.null
		las.Strt = tmp.strt
		las.Stop = tmp.stop
		las.Step = tmp.step
		las.Well = tmp.well
		curve := NewLasCurveFromString("SP.mV :spontaniously")
		las.Logs["SP"] = curve
		curve.Init(0, "SP", "SP", 5)
		las.SetNull(tmp.newNull)
		//os.Remove("empty.las")
		err := las.Save("~.las")
		assert.Nil(t, err)

		n, err := las.Open("~.las")
		assert.Equal(t, 0, n)
		assert.Equal(t, tmp.newNull, las.Null)
		assert.Equal(t, tmp.strt, las.Strt)
		assert.Equal(t, tmp.stop, las.Stop)
		assert.Equal(t, tmp.step, las.Step)
		assert.Equal(t, tmp.well, las.Well)
	}
	/*
		//стандартный las файл
		las := NewLas()
		las.Null = -99.99
		las.Strt = 0.201
		las.Stop = 10.01
		las.Step = 0.01
		las.Well = "Примерная-101/бис"
		curve := NewLasCurveFromString("SP.mV :spontaniously")
		las.Logs["SP"] = curve
		curve.Init(0, "SP", "SP", 5)
		las.SetNull(-0)
		//os.Remove("empty.las")
		err := las.Save("empty.las")
		assert.Nil(t, err)

		n, err := las.Open("empty.las")
		assert.Equal(t, 0, n)
		assert.Equal(t, -99.99, las.Null)
		assert.Equal(t, 0.201, las.Strt)
		assert.Equal(t, 10.01, las.Stop)
		assert.Equal(t, 0.01, las.Step)
		assert.Equal(t, "Примерная-101/бис", las.Well)

		// las файл в формате 866
		las = NewLas(cpd.CP866)
		las.Null = -99.99
		las.Strt = 0.201
		las.Stop = 10.01
		las.Step = 0.01
		las.Well = "Примерная-101/бис"
		curve = NewLasCurveFromString("SP.mV :spontaniously")
		las.Logs["SP"] = curve
		curve.Init(0, "SP", "SP", 5)
		las.SetNull(100.001)
		os.Remove("empty.las")
		err = las.Save("empty.las")
		assert.Nil(t, err)

		n, err = las.Open("empty.las")
		assert.Equal(t, 0, n)
		assert.Equal(t, 100.001, las.Null)
		assert.Equal(t, 0.201, las.Strt)
		assert.Equal(t, 10.01, las.Stop)
		assert.Equal(t, 0.01, las.Step)
		assert.Equal(t, "Примерная-101/бис", las.Well)
		os.Remove("empty.las")

		// las файл в формате UTF-8
		las = NewLas(cpd.UTF8)
		las.Null = -99.999
		las.Strt = 0.1
		las.Stop = 10.1
		las.Step = 0.1
		las.Well = "Примерная-101А/бис"
		curve = NewLasCurveFromString("SP.mV :spontaniously")
		las.Logs["SP"] = curve
		curve.Init(0, "SP", "SP", 5)
		las.SetNull(-999.25)
		//os.Remove("empty.las")
		err = las.Save("empty.las")
		assert.Nil(t, err)

		n, err = las.Open("empty.las")
		assert.Equal(t, 0, n)
		assert.Equal(t, -999.25, las.Null)
		assert.Equal(t, 0.1, las.Strt)
		assert.Equal(t, 10.1, las.Stop)
		assert.Equal(t, 0.1, las.Step)
		assert.Equal(t, "Примерная-101А/бис", las.Well)
		//os.Remove("empty.las")
	*/
}
