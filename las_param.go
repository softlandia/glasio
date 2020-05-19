package glasio

import (
	"fmt"
	"sort"
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

// ParseParamStr - parse string from las file
// return slice with 4 string and error if occure
// before process input string 2 or more space replace on 1 space
// sample "NULL .		               -9999.00 : Null value"
// f[0] - name
// f[1] - unit
// f[2] - value
// f[3] - description
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

const defCurveName = "-EL-" // curve name for null input

// ParseCurveStr - parse input string to 3 separated string
// " пс повт . мВ      : 7 кр сам"
//   ^^^^^^^   ^^        ^^^^^^^^
//   name      unit      description
// f[2] - description
// f[1] - unit
// f[0] - name
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
	if iDot < 0 { //if dot not found, all string is Curve name
		f[0] = strings.TrimSpace(s)
		return
	}
	f[0] = strings.TrimSpace(s[:iDot])
	if len(f[0]) == 0 {
		// case empty curve name
		f[0] = defCurveName
	}

	/*	s = strings.TrimSpace(s[iDot+1:])
		f[1], _ = xlib.StrCopyStop(s, ' ', ':') */
	f[1] = strings.TrimSpace(s[iDot+1:])
	return
}

//NewLasParam - create new object LasParam
//fill fields from s
func NewLasParam(s string) *LasParam {
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

//NULL .   -9999.00        : Null value
//WELL .   1 - Вообщевская :
// по умолчанию строка параметра разбирается на 4 составляющие: "имя параметра, ед измерения, значение, коментарий"
// между точкой и двоеточием ожидается единица измерения и значение параметра
// для параметра WELL пробел после точки также разбивает строку на две: ед измерения и значение
// но ТОЛЬКО для этого параметра единица измерения не существует и делать этого не следует
// таким образом собираем обратно в одно значение то, что ВОЗМОЖНО разбилось
func wellNameFromParam(p *LasParam) string {
	if len(p.Unit) == 0 {
		return p.Val
	}
	if len(p.Val) == 0 {
		return p.Unit //TODO не тестируется
	}
	return p.Unit + " " + p.Val
}

//LasCurve - class to store one log in Las
type LasCurve struct {
	LasParam
	Index int
	D     []float64
	V     []float64
}

// Cmp - compare current curve with another
// не сравниваются хранящиеся числовые данные (сам каротаж), только описание кривой, также не сравнивается индекс
// for deep comparison with all data points stored in the container use DeepCmp
func (o *LasCurve) Cmp(curve LasCurve) (res bool) {
	res = (o.LasParam == curve.LasParam)
	return
}

//SetLen - crop logs to actually len
//new len must be > 0 and < exist length
func (o *LasCurve) SetLen(n int) {
	if (n <= 0) || n >= len(o.D) {
		return
	}
	t := make([]float64, n)
	copy(t, o.D)
	o.D = t
	t = make([]float64, n)
	copy(t, o.V)
	o.V = t
}

// Init - initialize LasCurve, set index, name, mnemonic, make slice for store data
// перенести в NewLasCurve нельзя, в Init передаются обработанные данные кривой (словари, дублирование и т.д.)
func (o *LasCurve) Init(index int, mnemonic, name string, size int) {
	o.Index = index
	o.Mnemonic = mnemonic
	o.Name = name
	o.D = make([]float64, size)
	o.V = make([]float64, size)
}

//NewLasCurve - create new object LasCurve
func NewLasCurve(s string) LasCurve {
	lc := LasCurve{}
	curveFields := ParseCurveStr(s)
	lc.Name = curveFields[0]
	lc.IName = curveFields[0]
	lc.Unit = curveFields[1]
	lc.Desc = curveFields[2]
	lc.Index = 0
	return lc
}

// String - return LasCurve as string
func (o LasCurve) String() string {
	return fmt.Sprintf("[\n{\n\"IName\": \"%s\",\n\"Name\": \"%s\",\n\"Mnemonic\": \"%s\",\n\"Unit\": \"%s\",\"Val\": \"%s\",\n\"Desc\": \"%s\"\n}\n]", o.IName, o.Name, o.Mnemonic, o.Unit, o.Val, o.Desc)
}

// LasCurves - container for store all curves of las file
// .Cmp(curves *LasCurves) bool - compare two curves containers
type LasCurves map[string]LasCurve

// Text - return string represent all curves parameters: IName, Name, Unit etc
//TODO need realization
func (o LasCurves) Text() string {
	return "-"
}

// Cmp - compare current curves container with another
// сравниваются:
//   количество кривых в контейнере
//   два хеша от строк с именами всех кривых
func (o LasCurves) Cmp(curves LasCurves) (res bool) {
	res = (len(o) == len(curves))
	if res {
		curvesName := make([]string, 0, len(curves))
		for k := range o {
			curvesName = append(curvesName, k)
		}
		sort.Strings(curvesName)
		var sb strings.Builder
		for _, k := range curvesName {
			sb.WriteString(k)
		}
		h1 := xlib.StrHash(sb.String())
		curvesName = curvesName[:0]
		for k := range curves {
			curvesName = append(curvesName, k)
		}
		sort.Strings(curvesName)
		sb.Reset()
		for _, k := range curvesName {
			sb.WriteString(k)
		}
		h2 := xlib.StrHash(sb.String())
		res = (h1 == h2)
	}
	return res
}
