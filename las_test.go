//(c) softland 2019
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

type tReadWellParamStep struct {
	s string
	v float64
}

var dReadWellParamStep = []tReadWellParamStep{
	{"STEP.M            0.10 : dept step", 0.1},   //0
	{"STEP.M\t0.10                      ", 0.1},   //1
	{"STEP .M           0.10 : dept step", 0.1},   //2
	{"STEP . M          0.10 : dept step", 0.1},   //3
	{"STEP M  \t  \t  10.0 \t: dept step", 10},    //4
	{"STEP              10 :   dept step", 10},    //5 нет ед.изм.
	{"STEP.\t10          : dept  \t step", 10},    //6 нет ед.изм.
	{" STEP   . M       10.0 : dept step", 10},    //7
	{" STEP . M   \t  10.0   : dept step", 10},    //8
	{"\t STEP   M        10 : dept step ", 10},    //9
	{"STEP \t  M       10 :  dept step  ", 10},    //10
	{"STEP \t  M  \t\t 10 :  dept step  ", 10},    //11
	{"STEP.m              :   dept step ", 00},    //12 нет значения но есть ед.изм.
	{"STEP           0.113:    dept step", 0.113}, //13 нет ед.изм.
	{"STEP.\t0.999       : dept  \t step", 0.999}, //14 нет ед.изм.
}

func TestReadWellParam(t *testing.T) {
	las := NewLas()
	for i, tmp := range dReadWellParamStep {
		las.ReadWellParam(tmp.s)
		assert.Equal(t, las.Step, tmp.v, fmt.Sprintf("<ReadWellParam> on test %d return STEP: '%f' expect: '%f'\n", i, las.Step, tmp.v))
	}
}

func TestGetMnemonic(t *testing.T) {
	Mnemonic, err := LoadStdMnemonicDic(fp.Join("data/mnemonic.ini"))
	assert.Nil(t, err, fmt.Sprintf("load std mnemonic error: %v\n check out 'data\\mnemonic.ini'", err))

	VocDic, err := LoadStdVocabularyDictionary(fp.Join("data/dic.ini"))
	assert.Nil(t, err, fmt.Sprintf("load std vocabulary dictionary error: %v\n check out 'data\\dic.ini'", err))

	las := NewLas()
	mnemonic := las.GetMnemonic("1")
	assert.Equal(t, mnemonic, "-", fmt.Sprintf("<GetMnemonic> return '%s', expected '-'\n", mnemonic))

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

	las = nil
	las = NewLas()
	las.maxWarningCount = 100
	las.Open(fp.Join("data/more_20_warnings.las"))
	if las.Warnings.Count() != 62 {
		las.SaveWarning(fp.Join("data/more_20_warnings.wrn"))
		assert.Equal(t, las.Warnings.Count(), 62, fmt.Sprintf("<TestReachingMaxAmountWarnings> on read file data\\more_20_warnings.las warning count: %d expected 62\n", las.Warnings.Count()))
	}
}

/*
func TestLoadHeaderUtf(t *testing.T) {
	las := NewLas()
	las.iCodepage, _ = cpd.FileCodePageDetect(fp.Join("data/encodings_utf8wbom.las"))
	las.LoadHeader(fp.Join("data/encodings_utf8wbom.las"))
	assert.Equal(t, 1.2, las.Ver, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed VER: %f, expected %f", las.Ver, 1.2))
	assert.Equal(t, "NO", las.Wrap, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed WRAP: %s, expected %s", las.Wrap, "NO"))
	assert.Equal(t, 1670.0, las.Strt, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed STRT: %f, expected %f", las.Strt, 1670.0))
	assert.Equal(t, 1660.0, las.Stop, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed STOP: %f, expected %f", las.Stop, 1660.0))
	assert.Equal(t, -0.1250, las.Step, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed STEP: %f, expected %f", las.Step, -0.1250))
	assert.Equal(t, -999.250, las.Null, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed NULL: %f, expected %f", las.Null, -999.250))
	assert.Equal(t, "Скважина ºᶟᵌᴬń #12", las.Well, fmt.Sprintf("<LoadHeader> file 'encodings_utf8wbom.las' readed WELL: %s, expected %s", las.Well, "Скважина ºᶟᵌᴬń #12"))
}

func TestLoadHeaderUtf16le(t *testing.T) {
	las := NewLas()
	las.iCodepage, _ = cpd.FileCodePageDetect(fp.Join("data/encodings_utf16lebom.las"))
	las.LoadHeader(fp.Join("data/encodings_utf16lebom.las"))
	assert.Equal(t, 1.2, las.Ver, fmt.Sprintf("file 'encodings_utf16lebom.las' readed VER: %f, expected %f", las.Ver, 1.2))
	assert.Equal(t, "NO", las.Wrap, fmt.Sprintf("file 'encodings_utf16lebom.las' readed WRAP: %s, expected %s", las.Wrap, "NO"))
	assert.Equal(t, 1670.0, las.Strt, fmt.Sprintf("file 'encodings_utf16lebom.las' readed STRT: %f, expected %f", las.Strt, 1670.0))
	assert.Equal(t, 1660.0, las.Stop, fmt.Sprintf("file 'encodings_utf16lebom.las' readed STOP: %f, expected %f", las.Stop, 1660.0))
	assert.Equal(t, -0.1250, las.Step, fmt.Sprintf("file 'encodings_utf16lebom.las' readed STEP: %f, expected %f", las.Step, -0.1250))
	assert.Equal(t, -999.25, las.Null, fmt.Sprintf("file 'encodings_utf16lebom.las' readed NULL: %f, expected %f", las.Null, -999.25))
	assert.Equal(t, "ºᶟᵌᴬń BLOCK", las.Well, fmt.Sprintf("file 'encodings_utf16lebom.las' readed WELL: %s, expected %s", las.Well, "ºᶟᵌᴬń BLOCK"))
}*/

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
	{fp.Join("data/2.0/cp1251_2.0_based.las"), 2.0, "NO", 0.0, 39.9, 0.3, -999.25, "Примерная-1/бис(ё)"},
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1.0, 1.0, 0.1, -9999.00, "12-Сплошная"},
	{fp.Join("data/more_20_warnings.las"), 1.2, "NO", 0.0, 0.0, 1.0, -32768.0, "6"}, //in las file STEP=0.0 but this incorrect, LoadHeader replace STEP to actual from data
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1.0, 1.0, 0.1, -9999.0, "12-Сплошная"},
	{fp.Join("data/1.2/sample.las"), 1.2, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "ANY ET AL OIL WELL #12"},
	{fp.Join("data/2.0/sample_2.0.las"), 2.0, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "AAAAA_2"},
	{fp.Join("data/duplicate_step.las"), 1.2, "NO", 1670.0, 1660.0, -0.1200, -999.2500, "ANY ET AL OIL WELL #12"}, //duplicate_step.las contains two line with STEP:: STEP.M -0.1250: STEP.M -0.1200: using LAST parameter
	{fp.Join("data/encodings_utf8.las"), 1.2, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "ANY ºᶟᵌᴬń OIL WELL #12"},
}

func TestLoadHeader(t *testing.T) {
	var las *Las
	for _, tmp := range dLoadHeader {
		las = NewLas()
		las.iCodepage, _ = cpd.FileCodepageDetect(tmp.fn)
		f, _ := os.Open(tmp.fn)
		las.Reader, _ = cpd.NewReader(f)
		las.FileName = tmp.fn
		las.scanner = bufio.NewScanner(las.Reader)
		las.LoadHeader()
		assert.Equal(t, las.Ver, tmp.ver, fmt.Sprintf("<LoadHeader> file '%s' readed VER: %f, expected %f", las.FileName, las.Ver, tmp.ver))
		assert.Equal(t, las.Wrap, tmp.wrap, fmt.Sprintf("<LoadHeader> file '%s' readed WRAP: %s, expected %s", las.FileName, las.Wrap, tmp.wrap))
		assert.Equal(t, las.Strt, tmp.strt, fmt.Sprintf("<LoadHeader> file '%s' readed STRT: %f, expected %f", las.FileName, las.Strt, tmp.strt))
		assert.Equal(t, las.Stop, tmp.stop, fmt.Sprintf("<LoadHeader> file '%s' readed STOP: %f, expected %f", las.FileName, las.Stop, tmp.stop))
		assert.Equal(t, las.Step, tmp.step, fmt.Sprintf("<LoadHeader> file '%s' readed STEP: %f, expected %f", las.FileName, las.Step, tmp.step))
		assert.Equal(t, las.Null, tmp.null, fmt.Sprintf("<LoadHeader> file '%s' readed NULL: %f, expected %f", las.FileName, las.Null, tmp.null))
		assert.Equal(t, las.Well, tmp.well, fmt.Sprintf("<LoadHeader> file '%s' readed WELL: %s, expected %s", las.FileName, las.Well, tmp.well))
	}
	//test error case
	las, err := LoadLasHeader("not_exist_file.las") //file not exist
	assert.NotNil(t, err)
	assert.Nil(t, las)
	las, err = LoadLasHeader(fp.Join("data/utf-32be-bom.las")) //file not exist
	assert.NotNil(t, err)
	assert.Nil(t, las)
}

func TestLoadLasHeader(t *testing.T) {
	for _, tmp := range dLoadHeader {
		las, err := LoadLasHeader(tmp.fn)
		assert.Nil(t, err)
		assert.Equal(t, las.Ver, tmp.ver, fmt.Sprintf("<LoadHeader> file '%s' readed VER: %f, expected %f", las.FileName, las.Ver, tmp.ver))
		assert.Equal(t, las.Wrap, tmp.wrap, fmt.Sprintf("<LoadHeader> file '%s' readed WRAP: %s, expected %s", las.FileName, las.Wrap, tmp.wrap))
		assert.Equal(t, las.Strt, tmp.strt, fmt.Sprintf("<LoadHeader> file '%s' readed STRT: %f, expected %f", las.FileName, las.Strt, tmp.strt))
		assert.Equal(t, las.Stop, tmp.stop, fmt.Sprintf("<LoadHeader> file '%s' readed STOP: %f, expected %f", las.FileName, las.Stop, tmp.stop))
		assert.Equal(t, las.Step, tmp.step, fmt.Sprintf("<LoadHeader> file '%s' readed STEP: %f, expected %f", las.FileName, las.Step, tmp.step))
		assert.Equal(t, las.Null, tmp.null, fmt.Sprintf("<LoadHeader> file '%s' readed NULL: %f, expected %f", las.FileName, las.Null, tmp.null))
		assert.Equal(t, las.Well, tmp.well, fmt.Sprintf("<LoadHeader> file '%s' readed WELL: %s, expected %s", las.FileName, las.Well, tmp.well))
	}
	//test error case
	las, err := LoadLasHeader("--.--") //file not exist
	assert.NotNil(t, err)
	assert.Nil(t, las)
	las, err = LoadLasHeader(fp.Join("data/utf-32be-bom.las")) //file not exist
	assert.NotNil(t, err)
	assert.Nil(t, las)
}
