//(c) softland 2020
//softlandia@gmail.com

package glasio

import "fmt"

// CheckRes - результаты проверки, получааем из функции doCheck()
// если проверка не прошла, то res будет false и
// в warning будет положено предупреждение для логов
// в критических случаях err != nil и в себе содержит сообщение, при этом warning содержит соответствующее предупреждение для логов
// если err == nil то это не критичная проверка
type CheckRes struct {
	name    string
	warning TWarning
	err     error
	res     bool
}

func (cr CheckRes) String() string {
	return fmt.Sprintf("check name: %s, result: %v", cr.name, cr.res)
}

type doCheck func(chk Check, las *Las) CheckRes

// Check - конкретная проверка, обязан реализовать функцию doCheck()
type Check struct {
	name    string
	section string
	message string
	do      doCheck
}

// CheckResults - map с результатами всех проверок
type CheckResults map[string]CheckRes

func (crs CheckResults) nullWrong() bool {
	_, ok := crs["NULL"]
	return ok
}

func (crs CheckResults) stepWrong() bool {
	_, ok := crs["STEP"]
	_, ok2 := crs["STPU"]
	return ok || ok2
}

func (crs CheckResults) wrapWrong() bool {
	_, ok := crs["WRAP"]
	return ok
}

func (crs CheckResults) curvesWrong() bool {
	_, ok := crs["CURV"]
	return ok
}

func (crs CheckResults) strtStopWrong() bool {
	_, ok := crs["SSTP"]
	return ok
}

func (crs CheckResults) wellWrong() bool {
	_, ok := crs["WELL"]
	return ok
}

// isFatal - return true if CheckResults contains at least one check result with fatal error
func (crs CheckResults) isFatal() bool {
	for _, c := range crs {
		if c.err != nil {
			return true
		}
	}
	return false
}

// fatal - return first not nil error if CheckResults contains at least one check result with fatal error
func (crs CheckResults) fatal() error {
	for _, c := range crs {
		if c.err != nil {
			return c.err
		}
	}
	return nil
}

// Checker - ПРОВЕРЩИК, содержит в себе всех отдельных проверщиков,
// методом check() вызавает последовательно всех своих проверщиков,
// результаты отправляет в Logger
type Checker map[string]Check

// возвращает map в который сложены результаты проверки которые дали ошибку
// если проверка прошла безошибочно, то её в результатах не будет
// полученный map содержит только ошибки
func (c Checker) check(las *Las) CheckResults {
	res := make(CheckResults, 0)
	// key - имя проверки
	// chk - сам проверщик
	for key, chk := range c {
		r := chk.do(chk, las)
		if !r.res {
			res[key] = r
		}
	}
	return res
}

/****************************/

// NewStdChecker - создание нового ПРОВЕРЩИКА las файла.
// WRAP = ON
// section ~Curve is empty
// STEP == 0
// NULL == 0
// STRT == STOP
// WELL is empty
func NewStdChecker() Checker {
	return Checker{
		"WRAP": Check{"WRAP", "~V", "WRAP = ON", wrapCheck},
		"CURV": Check{"CURV", "~C", "Curve section is empty", curvesIsEmpty},
		"STEP": Check{"STEP", "~W", "STEP = 0", stepCheck},
		"STPU": Check{"STPU", "~W", "STEP not exist", stepExistCheck},
		"NULL": Check{"NULL", "~W", "NULL = 0", nullCheck},
		"SSTP": Check{"SSTP", "~W", "STRT = STOP", strtStop},
		"WELL": Check{"WELL", "~W", "WELL = ''", wellIsEmpty},
		"STRT": Check{"STRT", "~W", "STRT not exist", strtExistCheck},
		"STOP": Check{"STOP", "~W", "STOP not exist", stopExistCheck},
	}
}

// результат проверки возвращаем всегда с готовым варнингом и ошибкой,
// уже в дальнейшем, те проверки у которых res == true не вносятся в итоговый отчёт
func stepExistCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, "__WRN__ parameter STEP not exist"}, nil, !las.IsStepEmpty()}
}

func stopExistCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, "__WRN__ parameter STOP not exist"}, nil, !las.IsStopEmpty()}
}

func strtExistCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, "__WRN__ parameter STRT not exist"}, nil, !las.IsStrtEmpty()}
}

func wrapCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecVersion, las.currentLine, "__ERR__ WRAP = YES, file ignored"}, fmt.Errorf("Wrapped files not support"), !las.IsWraped()}
}

func curvesIsEmpty(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecCurInfo, las.currentLine, "__ERR__ Curve section is empty, file ignored"}, fmt.Errorf("Curve section not exist"), len(las.Logs) > 0}
}

func stepCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, fmt.Sprint("__WRN__ STEP parameter equal 0")}, nil, las.Step != 0.0}
}

func nullCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, fmt.Sprint("__WRN__ NULL parameter equal 0")}, nil, las.Null != 0.0}
}

func strtStop(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, fmt.Sprintf("__WRN__ STRT: %4.3f == STOP: %4.3f", las.Strt, las.Stop)}, nil, las.Strt != las.Stop}
}

func wellIsEmpty(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, las.currentLine, fmt.Sprintf("__WRN__ WELL: '%s' is empty", las.Well)}, nil, len(las.Well) != 0}
}
