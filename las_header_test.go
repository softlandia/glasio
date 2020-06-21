//(c) softland 2020
//softlandia@gmail.com
//test for HEADER section

package glasio

import (
	"fmt"
	fp "path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	{fp.Join("data/expand_points_01.las"), 1.2, "NO", 1.0, 1.0, 0.1, -9999.0, "12-Сплошная"},
	{fp.Join("data/1.2/sample.las"), 1.2, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "ANY ET AL OIL WELL #12"},
	{fp.Join("data/2.0/sample_2.0.las"), 2.0, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "AAAAA_2"},
	{fp.Join("data/broken_parameter.las"), 2.0, "NO", 0.0, 39.9, 0.3, -999.25, "Примерная-1 / бис(ё)"},            // file contain error on STRT
	{fp.Join("data/broken_header.las"), 2.0, "NO", 0.0, 39.9, 0.3, -999.25, "Примерная-1 / бис(ё)"},               // file contain only part of header
	{fp.Join("data/more_20_warnings.las"), 1.2, "NO", 0.0, 0.0, 0.0, -32768.0, "6"},                               // TODO STEP=0.0 but this incorrect, LoadHeader must replace STEP to actual from data
	{fp.Join("data/duplicate_step.las"), 1.2, "NO", 1670.0, 1660.0, -0.1200, -999.2500, "ANY ET AL OIL WELL #12"}, // duplicate_step.las contains two line with STEP:: STEP.M -0.1250: STEP.M -0.1200: using LAST parameter
	{fp.Join("data/encodings_utf8.las"), 1.2, "NO", 1670.0, 1660.0, -0.1250, -999.2500, "Скв #12Ω"},               // well name contain unicode char
	{fp.Join("test_files/warning-test-files/01-STOP-02.las"), 2.0, "NO", 0.0, -999.2500, 0.1, -999.25, "Скв #12"}, // STOP not exist
}

func TestLoadHeader(t *testing.T) {
	var las *Las
	for _, tmp := range dLoadHeader {
		las = NewLas()
		las.Open(tmp.fn)
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
