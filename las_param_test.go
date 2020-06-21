//(c) softland 2019
//softlandia@gmail.com

package glasio

import (
	"fmt"
	fp "path/filepath"
	"testing"

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

type tParseParamStr struct {
	s  string
	f0 string //str after PrepareParamStr
	f1 string
	f2 string
	f3 string
	f4 string
}

var dParseParamStr = []tParseParamStr{
	{"STEP.M            10 : dept step", "STEP.M 10:dept step", "STEP", "M", "10", "dept step"},           //0 'STEP.M 10 : dept step'
	{"STEP.M            10            ", "STEP.M 10", "STEP", "M", "10", ""},                              //1 "STEP.M 10 "
	{"STEP .M           10 : dept step", "STEP.M 10:dept step", "STEP", "M", "10", "dept step"},           //2 "STEP .M 10 : dept step"
	{"STEP . M          10 : dept step", "STEP.M 10:dept step", "STEP", "M", "10", "dept step"},           //3
	{"STEP M          10.0 : dept step", "STEP M 10.0:dept step", "STEP", "M", "10.0", "dept step"},       //4
	{"ST/M              10 : dept step", "ST/M 10:dept step", "ST/M", "10", "", "dept step"},              //5
	{"STEP              10 : dept step", "STEP 10:dept step", "STEP", "10", "", "dept step"},              //6
	{" STEP . M       10.0 : dept step", "STEP.M 10.0:dept step", "STEP", "M", "10.0", "dept step"},       //7
	{"\t STEP   M       10 :dept step ", "STEP M 10:dept step", "STEP", "M", "10", "dept step"},           //8
	{"ШАГ.M         0.0  :шаг глубины ", "ШАГ.M 0.0:шаг глубины", "ШАГ", "M", "0.0", "шаг глубины"},       //9
	{"ШАГ. M          :шаг по глубине ", "ШАГ.M:шаг по глубине", "ШАГ", "M", "", "шаг по глубине"},        //10
	{"ШАГ  M                          ", "ШАГ M", "ШАГ", "M", "", ""},                                     //11
	{"ШАГ  M          11              ", "ШАГ M 11", "ШАГ", "M", "11", ""},                                //12
	{"ШАГ             :шаг по глубине ", "ШАГ:шаг по глубине", "ШАГ", "", "", "шаг по глубине"},           //13
	{"ШАГ :m     0.2  :шаг по глубине ", "ШАГ:m 0.2:шаг по глубине", "ШАГ", "m", "0.2", "шаг по глубине"}, //14
	{"шаг.     : 2 сам             ", "шаг.:2 сам", "шаг", "", "", "2 сам"},                               //15
	{"шаг.m\t10.0     : 2 сам         ", "шаг.m 10.0:2 сам", "шаг", "m", "10.0", "2 сам"},                 //16
	{"STEP .m\t\t 10.0 \t: 2 сам      ", "STEP.m 10.0:2 сам", "STEP", "m", "10.0", "2 сам"},               //17
	{"VERS.             1.20: cp_866  ", "VERS.1.20:cp_866", "VERS", "1.20", "", "cp_866"},                //18
	{"NULL.   -999.250   :NULL VALUE", "NULL.-999.250:NULL VALUE", "NULL", "-999.250", "", "NULL VALUE"},  //19
	{"VERS.      2.0 :[Softland]", "VERS.2.0:[Softland]", "VERS", "2.0", "", "[Softland]"},                //20
	//{"WELL. Примерная 101/\"бис\" :well", "WELL.Примерная 101/\"бис\":well", "WELL", "Примерная 101/\"бис\"", "", "well"}, //21
}

func TestPrepareParamStr(t *testing.T) {
	for _, tmp := range dParseParamStr {
		s := PrepareParamStr(tmp.s)
		assert.Equal(t, tmp.f0, s)
	}
}

func TestParseParamStr(t *testing.T) {
	for _, tmp := range dParseParamStr {
		f := ParseParamStr(tmp.s)
		assert.Equal(t, tmp.f1, f[0])
		assert.Equal(t, tmp.f2, f[1])
		assert.Equal(t, tmp.f3, f[2])
		assert.Equal(t, tmp.f4, f[3])
	}
}

type tWellInfoStr struct {
	s  string
	f1 string
	f2 string
	f3 string
	f4 string
}

var dWellInfoStr = []tWellInfoStr{
	{"STEP.M            10 : dept step", "STEP", "M", "10", "dept step"},       //0
	{"STEP.M            10            ", "STEP", "M", "10", ""},                //1
	{"STEP .M           10 : dept step", "STEP", "M", "10", "dept step"},       //2
	{"STEP . M          10 : dept step", "STEP", "M", "10", "dept step"},       //3
	{"STEP M          10.0 : dept step", "STEP", "M", "10.0", "dept step"},     //4
	{"ST/M              10 : dept step", "ST/M", "", "10", "dept step"},        //5
	{"STEP              10 : dept step", "STEP", "", "10", "dept step"},        //6
	{" STEP . M       10.0 : dept step", "STEP", "M", "10.0", "dept step"},     //7
	{"\t STEP   M       10 :dept step ", "STEP", "M", "10", "dept step"},       //8
	{"ШАГ.M         0.0  :шаг глубины ", "ШАГ", "M", "0.0", "шаг глубины"},     //9
	{"ШАГ. M          :шаг по глубине ", "ШАГ", "", "M", "шаг по глубине"},     //10
	{"ШАГ  M                          ", "ШАГ", "", "M", ""},                   //11
	{"ШАГ  M          11              ", "ШАГ", "M", "11", ""},                 //12
	{"ШАГ             :шаг по глубине ", "ШАГ", "", "", "шаг по глубине"},      //13
	{"ШАГ :m     0.2  :шаг по глубине ", "ШАГ", "m", "0.2", "шаг по глубине"},  //14
	{"шаг.     : 2 сам                ", "шаг", "", "", "2 сам"},               //15
	{"шаг.m\t10.0     : 2 сам         ", "шаг", "m", "10.0", "2 сам"},          //16
	{"STEP .m\t\t 10.0 \t: 2 сам      ", "STEP", "m", "10.0", "2 сам"},         //17
	{"VERS.             1.20: cp_866  ", "VERS", "", "1.20", "cp_866"},         //18
	{"NULL.   -999.250     :NULL VALUE", "NULL", "", "-999.250", "NULL VALUE"}, //19
	{"VERS.      2.0 :[Softland]      ", "VERS", "", "2.0", "[Softland]"},      //20
}

func TestNewLasParamFromString(t *testing.T) {
	var lp *LasParam
	for _, tmp := range dWellInfoStr {
		lp = NewLasParam(tmp.s)
		assert.Equal(t, tmp.f1, lp.Name)
		assert.Equal(t, tmp.f2, lp.Unit)
		assert.Equal(t, tmp.f3, lp.Val)
		assert.Equal(t, tmp.f4, lp.Desc)
		lp = nil
	}
}

type tParseCurveStr struct {
	s  string
	f0 string
	f1 string
	f2 string
}

var dParseCurveStr = []tParseCurveStr{
	{" ПС   мВ.", "ПС мВ", "", ""},                                //12
	{"ПС.мВ          : 1 кр сам", "ПС", "мВ", "1 кр сам"},         //1
	{"ПС.           : 2 кр сам ", "ПС", "", "2 кр сам"},           //2
	{"ПС повторная  :3  кр  сам", "ПС повторная", "", "3 кр сам"}, //3
	{" ПС  \t  \t    : 4 кр сам", "ПС", "", "4 кр сам"},           //4
	{" ПС            :         ", "ПС", "", ""},                   //5
	{" ПС . мВ       : 6 кр сам", "ПС", "мВ", "6 кр сам"},         //6
	{" пс повт . мВ  : 7 кр сам", "пс повт", "мВ", "7 кр сам"},    //7
	{" ПС \t  мВ     : 8 кр сам", "ПС мВ", "", "8 кр сам"},        //8
	{" ПС   . мВ               ", "ПС", "мВ", ""},                 //9
	{" ПС   повт               ", "ПС повт", "", ""},              //10
	{" ПС   мВ      : 11 кр сам", "ПС мВ", "", "11 кр сам"},       //11
	{" .      : ", "-EL-", "", ""},                                //12
	{" .mv      : ", "-EL-", "mv", ""},                            //13
	{" .mv      :sp", "-EL-", "mv", "sp"},                         //14
	{" .m v     :sp", "-EL-", "m v", "sp"},                        //15
}

func TestParseCurveStr(t *testing.T) {
	for _, tmp := range dParseCurveStr {
		f := ParseCurveStr(tmp.s)
		assert.Equal(t, tmp.f0, f[0])
		assert.Equal(t, tmp.f1, f[1])
		assert.Equal(t, tmp.f2, f[2])
	}
}

func TestNewLasCurveFromString(t *testing.T) {
	las := NewLas()
	for _, tmp := range dParseCurveStr {
		lc := NewLasCurve(tmp.s, las)
		assert.Equal(t, tmp.f0, lc.Name)
		assert.Equal(t, tmp.f1, lc.Unit)
		assert.Equal(t, tmp.f2, lc.Desc)
	}
}

func TestLasCurveSetLen(t *testing.T) {
	las := NewLas()
	curve := NewLasCurve("SP.mV  :self", las)
	//curve.Init(0, "SP", "SP", 5)
	curve.D = append(curve.D, 0.1, 0.2, 0.3, 0.4, 0.5)
	curve.SetLen(3)
	assert.Equal(t, 3, len(curve.D))
	assert.Equal(t, 3, len(curve.V))
	assert.Equal(t, 0.3, curve.D[2])

	curve.SetLen(4) //nothing to do, size of data slice not change
	assert.Equal(t, 3, len(curve.D))
	assert.Equal(t, 3, len(curve.V))

	curve.SetLen(0) //nothing to do, size of data slice not change
	assert.Equal(t, 3, len(curve.D))
	assert.Equal(t, 3, len(curve.V))

	curve.SetLen(-5) //nothing to do, size of data slice not change
	assert.Equal(t, 3, len(curve.D))
	assert.Equal(t, 3, len(curve.V))

	curve.SetLen(2)
	assert.Equal(t, 2, len(curve.D))
	assert.Equal(t, 2, len(curve.V))
	assert.Equal(t, 0.2, curve.D[1])
}

// Тестирование отображения кривой в строковое представление
func TestLasCurveString(t *testing.T) {
	las := makeLasFromFile(fp.Join("data/test-curve-sec-empty-mnemonic.las"))
	//D.M     : 1  DEPTH
	assert.Equal(t, "[\n{\n\"IName\": \"D\",\n\"Name\": \"D\",\n\"Mnemonic\": \"\",\n\"Unit\": \"M\",\"Val\": \"\",\n\"Desc\": \"1 DEPTH\"\n}\n]", las.Logs[0].String())
	//A.US/M  : 2  SONIC TRANSIT TIME
	assert.Equal(t, "[\n{\n\"IName\": \"A\",\n\"Name\": \"A\",\n\"Mnemonic\": \"\",\n\"Unit\": \"US/M\",\"Val\": \"\",\n\"Desc\": \"2 SONIC TRANSIT TIME\"\n}\n]", las.Logs[1].String())
	//-EL-1.      :
	assert.Equal(t, "[\n{\n\"IName\": \"-EL-\",\n\"Name\": \"-EL-\",\n\"Mnemonic\": \"\",\n\"Unit\": \"\",\"Val\": \"\",\n\"Desc\": \"\"\n}\n]", las.Logs[4].String())
	assert.Equal(t, "[\n{\n\"IName\": \"-EL-\",\n\"Name\": \"-EL-5\",\n\"Mnemonic\": \"\",\n\"Unit\": \"m\",\"Val\": \"\",\n\"Desc\": \"\"\n}\n]", las.Logs[5].String())
	assert.Equal(t, "[\n{\n\"IName\": \"-EL-\",\n\"Name\": \"-EL-6\",\n\"Mnemonic\": \"\",\n\"Unit\": \"v/v\",\"Val\": \"\",\n\"Desc\": \"\"\n}\n]", las.Logs[6].String())
	assert.Equal(t, "[\n{\n\"IName\": \"-EL-\",\n\"Name\": \"-EL-7\",\n\"Mnemonic\": \"\",\n\"Unit\": \"m V\",\"Val\": \"\",\n\"Desc\": \"\"\n}\n]", las.Logs[7].String())
	assert.Equal(t, "[\n{\n\"IName\": \"-EL-5\",\n\"Name\": \"-EL-58\",\n\"Mnemonic\": \"\",\n\"Unit\": \"m\",\"Val\": \"\",\n\"Desc\": \"\"\n}\n]", las.Logs[8].String())
}
