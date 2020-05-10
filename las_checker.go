//(c) softland 2020
//softlandia@gmail.com

package glasio

import "fmt"

// CheckRes - результаты проверки, получааем из функции doCheck()
// если проверка не прошла, то res будет false и
// в warning будет положено предупреждение для логов
// в критических случаях err содержит и ошибку и warning, предупреждение для одних логов, а ошибка для прекращения
// если err == nil то это не критичная проверка
type CheckRes struct {
	name    string
	warning TWarning
	err     error
	res     bool
	/*
		section string
		message string
	*/
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

// CheckResults - слайс с результатами всех проверок
type CheckResults map[string]CheckRes

func (crs CheckResults) nullWrong() bool {
	_, ok := crs["NULL"]
	return ok
}

func (crs CheckResults) stepWrong() bool {
	_, ok := crs["STEP"]
	return ok
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

// Checker - ПРОВЕРЩИК, содержит в себе всех отдельных проверщиков,
// методом check() вызавает последовательно всех своих проверщиков,
// результаты отправляет в Logger
type Checker map[string]Check

func (c Checker) check(las *Las) CheckResults {
	res := make(CheckResults, 0)
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
		"NULL": Check{"NULL", "~W", "NULL = 0", nullCheck},
		"SSTP": Check{"SSTP", "~W", "STRT = STOP", strtStop},
		"WELL": Check{"WELL", "~W", "WELL = ''", wellIsEmpty},
	}
}

// результат проверки возвращаем всегда с готовым варнингом и ошибкой,
// уже в дальнейшем, те проверки у которых res == true не вносятся в итоговый отчёт
func wrapCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecVersion, -1, "__ERR__ WRAP = YES, file ignored"}, fmt.Errorf("Wrapped files not support"), !las.IsWraped()}
}

func curvesIsEmpty(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecCurInfo, -1, "__ERR__ Curve section is empty, file ignored"}, fmt.Errorf("Curve section not exist"), len(las.Logs) > 0}
}

func stepCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprint("__WRN__ STEP parameter equal 0")}, nil, las.Step != 0.0}
}

func nullCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprint("__WRN__ NULL parameter equal 0")}, nil, las.Null != 0.0}
}

func strtStop(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("__WRN__ STRT: %4.3f == STOP: %4.3f", las.Strt, las.Stop)}, nil, las.Strt != las.Stop}
}

func wellIsEmpty(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("__WRN__ WELL: '%s' is empty", las.Well)}, nil, len(las.Well) != 0}
}
