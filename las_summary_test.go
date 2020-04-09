//(c) softland 2020
//softlandia@gmail.com
package glasio

import (
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
	curv int //количество кривых в файле
	nums int //количество точек в файле
	werr bool
}

var dSummaryCheck = []tSummaryCheck{
	{fp.Join("data/6038187_v1.2.las"), 2.0, "NO", 0.05, 136.6, 0.05, -99999, "Scorpio E1", 9, 2732, false},
	{fp.Join("data/6038187_v1.2_short.las"), 2.0, "NO", 0.05, 136.6, 0.05, -99999, "Scorpio E1", 9, 121, false},
	{fp.Join("data/1001178549.las"), 2.0, "YES", 1783.5, 1784.5, 0.25, -999.25, "1-28", 27, 0, true},
	{fp.Join("data/alog.las"), 1.20, "NO", 0, 0, 0.05, -999.25, "", 9, 24, false},
	{fp.Join("data/autodepthindex_F.las"), 1.20, "NO", 0, 100, 1, -999.25, "ANY ET AL OIL WELL #12", 2, 101, false},
	{fp.Join("data/barebones.las"), 2.0, "NO", 0, 0, 0, 0, "", 1, 0, true},
	{fp.Join("data/barebones2.las"), 2.0, "NO", 0, 0, -0.1, -999.25, "", 0, 0, true},
	{fp.Join("data/blank_line.las"), 2.0, "NO", -999.25, -999.25, 0.0833333333333333, -999.25, "", 1, 0, true},
	{fp.Join("data/data_characters.las"), 2.0, "NO", 0, 0, 10, -999.25, "", 4, 0, true},
	{fp.Join("data/duplicate_step.las"), 1.2, "NO", 1670, 1660, -0.12, -999.25, "ANY ET AL OIL WELL #12", 8, 3, false},
	{fp.Join("data/encodings_utf8.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf8_20.las"), 2.0, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf8wbom.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скважина #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf8wbom_20.las"), 2.0, "NO", 1670, 1660, -0.125, -999.25, "Скважина #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf16be.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf16bebom.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf16le.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false},
	{fp.Join("data/encodings_utf16lebom.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "Скв #12Ω", 8, 3, false},
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1, 1, 0.1, -9999.00, "12-Сплошная", 4, 7, false},
	{fp.Join("data/logging_levels.las"), 2.0, "NO", 0, 7273.5, 0.25, -999.25, "TOTEM # 9", 18, 29095, false},
	{fp.Join("data/missing_null.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, false},
	{fp.Join("data/missing_vers.las"), 2.0, "NO", 1670, 1660, -0.125, -999.25, "WELL", 8, 3, false},
	{fp.Join("data/missing_wrap.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, false},
	{fp.Join("data/more_20_warnings.las"), 1.2, "NO", 0.0, 0.0, 1.0, -32768.0, "6", 6, 23, true}, //in las file STEP=0.0 but this incorrect, LoadHeader replace STEP to actual from data
	{fp.Join("data/no-data-section.las"), 1.2, "NO", 0.0, 0.0, -32768, -32768.0, "6", 31, 0, true},
	{fp.Join("data/sample_bracketed_units.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, true},
	{fp.Join("data/test-curve-sec-empty-mnemonic.las"), 1.2, "NO", 1670, 1669.75, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, true},
	{fp.Join("data/UWI_API_leading_zero.las"), 1.2, "NO", 1670, 1660, -0.125, -999.25, "ANY ET AL OIL WELL #12", 8, 3, true},
}

// Основной тест по массиву готовых las файлов
func TestSummaryRead(t *testing.T) {
	for _, tmp := range dSummaryCheck {
		las := NewLas()
		n, _ := las.Open(tmp.fn)
		assert.Equal(t, tmp.nums, n)
		assert.Equal(t, tmp.curv, len(las.Logs))
		assert.Equal(t, tmp.ver, las.Ver)
		assert.Equal(t, tmp.wrap, las.Wrap)
		assert.Equal(t, tmp.strt, las.Strt)
		assert.Equal(t, tmp.stop, las.Stop)
		assert.Equal(t, tmp.step, las.Step)
		assert.Equal(t, tmp.null, las.Null)
		assert.Equal(t, tmp.well, las.Well)
	}
}

func TestCurveSec1(t *testing.T) {
	las := NewLas()
	n, err := las.Open(fp.Join("data/test-curve-sec-empty-mnemonic.las"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, "D", las.Logs["D"].Name)
	assert.Equal(t, "M", las.Logs["D"].Unit)
}

func TestCurveSec2(t *testing.T) {
	correct := makeLasFromFile(fp.Join("data/test-curve-sec-empty-mnemonic+.las"))
	las := makeLasFromFile(fp.Join("data/test-curve-sec-empty-mnemonic.las"))
	assert.True(t, cmpLas(correct, las))
	assert.True(t, correct.Logs.Cmp(las.Logs))
	las = makeLasFromFile(fp.Join("data/missing_wrap.las"))
	assert.False(t, cmpLas(correct, las))
	assert.False(t, correct.Logs.Cmp(las.Logs))
}
