package glasio

import (
	"strings"

	"github.com/softlandia/xlib"
)

//LasParam - class to store parameter from any section
type LasParam struct {
	IName    string
	Name     string
	Mnemonic string
	Unit     string
	Val      string
	Desc     string
}

//PrepareParamStr - prepare string to parse, replace many space to one, replace tab to space, replace combination of separator to one
func PrepareParamStr(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = xlib.ReplaceAllSpace(s)
	s = xlib.ReplaceSeparators(s)
	return strings.TrimSpace(s)
}

//ParseParamStr - parse string from las file
//return slice with 4 string and error if occure
//before process input string 2 or more space replace on 1 space
func ParseParamStr(s string) (f [4]string) {
	f[0] = ""
	f[1] = ""
	f[2] = ""
	f[3] = ""
	s = PrepareParamStr(s)

	iComma := strings.LastIndex(s, ":") //comment parse first, cut string after
	commentFlag := (iComma >= 0)
	if commentFlag {
		f[3] = s[iComma+1:]
		s = strings.TrimSpace(s[:iComma])
	}

	var iDot int
	f[0], iDot = xlib.StrCopyStop(s, ' ', '.', ':')
	f[0] = strings.TrimSpace(f[0])
	if iDot >= len(s) {
		return
	}
	s = strings.TrimSpace(s[iDot+1:])
	f[1], iDot = xlib.StrCopyStop(s, ' ', ':')
	f[1] = strings.TrimSpace(f[1])
	if iDot >= len(s) {
		return
	}
	s = strings.TrimSpace(s[iDot+1:])
	f[2], _ = xlib.StrCopyStop(s, ':')
	f[2] = strings.TrimSpace(f[2])
	return
}

//ParseCurveStr - parse input string to 3 separated string
//" пс повт . мВ  : 7 кр сам"
func ParseCurveStr(s string) (f [3]string) {
	f[0] = ""
	f[1] = ""
	f[2] = ""
	s = PrepareParamStr(s)

	iComma := strings.LastIndex(s, ":") //comment parse first, cut string after
	commentFlag := (iComma >= 0)
	if commentFlag {
		f[2] = s[iComma+1:]
		s = strings.TrimSpace(s[:iComma])
	}
	//if comma not found, string not change
	//now s contains only name and unit

	iDot := strings.Index(s, ".")
	//TODO если в строке нет точки, то отделить имя от единицы измерения невозможно. имя может содержать произвольные символя включая пробел
	if iDot < 0 { //if dot not found, all string is Curve name
		f[0] = strings.TrimSpace(s)
		return
	}
	f[0] = strings.TrimSpace(s[:iDot])

	s = strings.TrimSpace(s[iDot+1:])
	f[1], _ = xlib.StrCopyStop(s, ' ', ':')
	f[1] = strings.TrimSpace(f[1])
	return
}

//NewLasParamFromString - create new object LasParam
//fill fields from s
func NewLasParamFromString(s string) *LasParam {
	par := new(LasParam)
	paramFields := ParseParamStr(s)
	par.Name = paramFields[0]
	par.Unit = paramFields[1]
	par.Val = paramFields[2]
	if (len(par.Val) == 0) && (len(par.Unit) > 0) {
		par.Val = par.Unit
		par.Unit = ""
	}
	par.Desc = paramFields[3]
	return par
}

//LasCurve - class to store one log in Las
type LasCurve struct {
	LasParam
	Index int
	dept  []float64
	log   []float64
}

//SetLen - crop logs to actually len
//new len must be > 0 and < exist length
func (o *LasCurve) SetLen(n int) {
	if (n <= 0) || n >= len(o.dept) {
		return
	}
	t := make([]float64, n, n)
	copy(t, o.dept)
	o.dept = t
	t = make([]float64, n, n)
	copy(t, o.log)
	o.log = t
}

//Init - initialize LasCurve, set index, name, mnemonic, make slice for store data
func (o *LasCurve) Init(index int, mnemonic, name string, size int) {
	o.Index = index
	o.Mnemonic = mnemonic
	o.Name = name
	o.dept = make([]float64, size)
	o.log = make([]float64, size)
}

//NewLasCurveFromString - create new object LasCurve
func NewLasCurveFromString(s string) LasCurve {
	lc := LasCurve{}
	curveFields := ParseCurveStr(s)
	lc.Name = curveFields[0]
	lc.IName = curveFields[0]
	lc.Unit = curveFields[1]
	lc.Desc = curveFields[2]
	lc.Index = 0
	return lc
}
