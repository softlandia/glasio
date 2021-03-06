// (c) softland 2020
// softlandia@gmail.com
package glasio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	fp "path/filepath"
)

type tSummaryCheck struct {
	fn   string
	ver  float64
	wrap string
	strt float64
	stop float64
	step float64
	null float64
	well string
	curv int  //количество кривых в файле
	nums int  //количество точек в файле
	werr bool //не используется
	d1,  //значение первой точки по глубине
	dn, //значение последней точки по глубине (если есть)
	v1, //значение первой кривой на первой глубине
	vn float64 //значение первой кривой на последней глубине (если есть)
}

var dSummaryCheck = []tSummaryCheck{
	{fp.Join("data/barebones.las"), 2.0, "NO", 200, -999.25, 1.1, -999.25, "", 1, 2, true, 200.0, 201.1, 0, 0},
	{fp.Join("data/6038187_v1.2.las"), 2.0, "NO", 0.05, 136.6, 0.05, -99999, "Scorpio E1", 9, 2732, false, 0.05, 136.6, 49.7650, -56.2750},
	{fp.Join("data/6038187_v1.2_short.las"), 2.0, "NO", 0.05, 136.6, 0.05, -99999, "Scorpio E1", 9, 121, false, 12.0, 18.0, 101.78, 101.259},
	{fp.Join("data/1001178549.las"), 2.0, "YES", 1783.5, 1784.5, 0.25, -999.25, "1-28", 27, 0, true, 0, 0, 0, 0},
	{fp.Join("data/alog.las"), 1.20, "NO", 0, 0, 0.05, -999.25, "", 9, 24, false, 0.00, 0.00, 0.00, 0.00},
	{fp.Join("data/autodepthindex_F.las"), 1.20, "NO", 0, 100, 1, -999.25, "ANY ET AL OIL WELL #12", 2, 101, false, 0, 100, 0.730568506467, 0.959183036405},
	{fp.Join("data/barebones2.las"), 2.0, "NO", -999.25, -999.25, -999.25, -999.25, "", 0, 0, true, 0, 0, 0, 0}, // step и null не правятся, отсутствует секция Curve, ошибка заголовка
	{fp.Join("data/blank_line.las"), 2.0, "NO", -999.25, -999.25, 0.0833333333333333, -999.25, "", 1, 0, true, 0, 0, 0, 0},
	{fp.Join("data/data_characters.las"), 2.0, "NO", 0, 0, 10, -999.25, "", 4, 0, true, -999.25, -999.25, -999.25, -999.25},
	{fp.Join("data/duplicate_step.las"), 1.2, "NO", 1670, 1660, -0.125, -999.2111, "ANY ET AL OIL WELL #12", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf8.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf8_20.las"), 2.0, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf8wbom.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скважина #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf8wbom_20.las"), 2.0, "NO", 1670, 1660, -0.125, -999.25, "Скважина #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf16be.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf16bebom.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf16le.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/encodings_utf16lebom.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1, 1, 0.1, -9999.00, "12-Сплошная", 4, 7, false, 1.0, 1.6, -9999.0, 0},
	{fp.Join("data/logging_levels.las"), 2.0, "NO", 0, 7273.5, 0.25, -999.25, "TOTEM # 9", 18, 29095, false, 0.00, 7273.5, 1604.8491, -999.25},
	{fp.Join("data/missing_null.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, false, 1670.0, 1669.75, -999.25, 123.45},
	{fp.Join("data/missing_vers.las"), 2.0, "NO", 1670, 1660, -0.125, -999.25, "WELL", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/missing_wrap.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, false, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/more_20_warnings.las"), 1.2, "NO", 0.0, 0.0, 1.0, -32768.0, "6", 6, 22, true, 1, 2.2e11, -32768.0, 186}, //in file STEP=0.0 but this incorrect, LoadHeader replace STEP to actual from data
	{fp.Join("data/no-data-section.las"), 1.2, "NO", 0.0, 0.0, -32768.0, -32768.0, "6", 31, 0, true, 0, 0, 0, 0},           //in file STEP=0.0 but this incorrect, data section contain incorrect step too, result step equal NULL
	{fp.Join("data/sample_bracketed_units.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, true, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/test-curve-sec-empty-mnemonic.las"), 1.2, "NO", 1670, 1669.75, -0.125, -999.25, "ANY ET AL OIL WELL #12", 9, 3, true, 1670.0, 1669.75, 123.45, 123.45},
	{fp.Join("data/UWI_API_leading_zero.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, true, 1670.0, 1669.75, 123.45, 123.45},
}

// Основной тест по массиву готовых las файлов
// проверяются основные параметры заголовка, количество считанных точек, количество считанных кривых и 4 точки данных (две точки глубины и две точки данных)
func TestSummaryRead(t *testing.T) {
	for _, tmp := range dSummaryCheck {
		las := NewLas()
		n, _ := las.Open(tmp.fn)
		assert.Equal(t, tmp.nums, n, fmt.Sprintf("<TestSummaryRead> nums fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.ver, las.VERS(), fmt.Sprintf("<TestSummaryRead> ver fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.wrap, las.WRAP(), fmt.Sprintf("<TestSummaryRead> wrap fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.strt, las.STRT(), fmt.Sprintf("<TestSummaryRead> strt fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.stop, las.STOP(), fmt.Sprintf("<TestSummaryRead> stop fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.step, las.STEP(), fmt.Sprintf("<TestSummaryRead> step fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.null, las.NULL(), fmt.Sprintf("<TestSummaryRead> null fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.well, las.WELL(), fmt.Sprintf("<TestSummaryRead> null fail on file: '%s'\n", tmp.fn))
		//две проверки касающиеся секции кривых
		//кривые попадают в два контейнера: las.Logs и las.CurSec.params
		assert.Equal(t, tmp.curv, len(las.Logs), fmt.Sprintf("<TestSummaryRead> curves fail on file: '%s'\n", tmp.fn))
		assert.Equal(t, tmp.curv, len(las.CurSec.params), fmt.Sprintf("<TestSummaryRead> curves fail on file: '%s'\n", tmp.fn))
		//проверки по данным
		if tmp.nums > 0 {
			assert.Equal(t, tmp.d1, las.Logs[0].D[0], fmt.Sprintf("<TestSummaryRead> curves fail on file: '%s'\n", tmp.fn))
			assert.Equal(t, tmp.dn, las.Logs[0].D[n-1], fmt.Sprintf("<TestSummaryRead> curves fail on file: '%s'\n", tmp.fn))
			if tmp.curv > 1 {
				assert.Equal(t, tmp.v1, las.Logs[1].V[0], fmt.Sprintf("<TestSummaryRead> curves fail on file: '%s'\n", tmp.fn))
				assert.Equal(t, tmp.vn, las.Logs[1].V[n-1], fmt.Sprintf("<TestSummaryRead> curves fail on file: '%s'\n", tmp.fn))
			}
		}
	}
}

//Убрать тест
func TestCurveSec1(t *testing.T) {
	las := NewLas()
	n, err := las.Open(fp.Join("data/test-curve-sec-empty-mnemonic.las"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, "D", las.Logs[0].Name)
	assert.Equal(t, "M", las.Logs[0].Unit)
}

func TestCmpLas(t *testing.T) {
	correct := makeLasFromFile(fp.Join("data/test-curve-sec-empty-mnemonic+.las"))
	las := makeLasFromFile(fp.Join("data/test-curve-sec-empty-mnemonic.las"))
	assert.True(t, cmpLas(correct, las))
	assert.True(t, correct.Logs.Cmp(las.Logs))
	las = makeLasFromFile(fp.Join("data/missing_wrap.las"))
	assert.False(t, cmpLas(correct, las))
	assert.False(t, correct.Logs.Cmp(las.Logs))
}

func TestTabulatedData(t *testing.T) {
	correct := makeLasFromFile(fp.Join("data/tabulated_data+.las"))
	las := makeLasFromFile(fp.Join("data/tabulated_data.las"))
	assert.True(t, cmpLas(correct, las))
	assert.True(t, correct.Logs.Cmp(las.Logs))
	l := las.Logs[3]
	assert.Equal(t, 0.451, l.V[1])
	assert.Equal(t, "NPHI", l.Name)
}

func TestLasCheck(t *testing.T) {
	lasLog := LasCheck(fp.Join("data/test-curve-sec-empty-mnemonic+.las"))
	assert.NotNil(t, lasLog)
	s := lasLog.msgOpen.ToString()
	assert.Empty(t, s)
	s = lasLog.msgCheck.String()
	assert.Empty(t, s)
	s = lasLog.msgCurve.String(fp.Join("data/test-curve-sec-empty-mnemonic+.las"))
	assert.Contains(t, s, "test-curve-sec-empty-mnemonic+.las'##")

	// проверка на возврат nil при невозможности выполнить проверку
	lasLog, err := LasDeepCheck(fp.Join("data/test-curve-sec-empty-mnemonic+.las"), fp.Join("data/mnemonic.-"), fp.Join("data/dic.ini"))
	assert.Nil(t, lasLog)
	assert.NotNil(t, err)
	lasLog, err = LasDeepCheck(fp.Join("data/test-curve-sec-empty-mnemonic+.las"), fp.Join("data/mnemonic.ini"), fp.Join("data/dic.-"))
	assert.Nil(t, lasLog)
	assert.NotNil(t, err)

	lasLog, err = LasDeepCheck(fp.Join("data/test-curve-sec-empty-mnemonic+.las"), fp.Join("data/mnemonic.ini"), fp.Join("data/dic.ini"))
	assert.Nil(t, err)
	s = lasLog.msgCurve.String(fp.Join("data/test-curve-sec-empty-mnemonic+.las"))
	assert.Contains(t, s, "*input log: B 	 mnemonic:*")
	assert.Contains(t, s, "input log: SP 	 mnemonic: SP")
	assert.Contains(t, s, "*input log: -EL-58 	 mnemonic:*")
	s = lasLog.missMnemonic.String()
	assert.Contains(t, s, "-EL-7")
	assert.NotContains(t, s, "SP")

	lasLog, err = LasDeepCheck(fp.Join("data/more_20_warnings.las"), fp.Join("data/mnemonic.ini"), fp.Join("data/dic.ini"))
	assert.NotNil(t, lasLog)
	assert.Nil(t, err)
	s = lasLog.msgOpen.ToString()
	assert.Contains(t, s, "STEP parameter equal 0")
	assert.Contains(t, s, "__WRN__ STRT: 0.000 == STOP: 0.000")
	s = lasLog.msgCheck.String()
	assert.Empty(t, s)
	s = lasLog.msgCurve.String(fp.Join("data/more_20_warnings.las"))
	assert.Contains(t, s, "*input log: второй каротаж 	 mnemonic:*")
	assert.Contains(t, s, "input log: GK 	 mnemonic: GR")
	assert.Contains(t, s, "*input log: первый 	 mnemonic:*")
	s = lasLog.missMnemonic.String()
	assert.Contains(t, s, "NNB")
	assert.Contains(t, s, "второй каротаж")
	assert.NotContains(t, s, "GR")

	// случай если файла нет "data/-.las"
	lasLog, err = LasDeepCheck(fp.Join("data/-.las"), fp.Join("data/mnemonic.ini"), fp.Join("data/dic.ini"))
	assert.NotNil(t, lasLog)
	assert.NotNil(t, err)
	assert.NotNil(t, lasLog.errorOnOpen)

	// случай если las файл WRAP
	lasLog = LasCheck(fp.Join("data/1.2/sample_wrapped.las"))
	assert.Contains(t, lasLog.msgCheck.String(), "WRAP=YES")
	lasLog, _ = LasDeepCheck(fp.Join("data/1.2/sample_wrapped.las"), fp.Join("data/mnemonic.ini"), fp.Join("data/dic.ini"))
	assert.Contains(t, lasLog.msgCheck.String(), "WRAP=YES")
}
