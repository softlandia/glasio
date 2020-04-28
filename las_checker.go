//(c) softland 2020
//softlandia@gmail.com

package glasio

import "fmt"

// CheckRes - результаты проверки, получааем из функции doCheck()
type CheckRes struct {
	name    string
	section string
	message string
	err     error
	res     bool
}

func (cr CheckRes) String() string {
	return fmt.Sprintf("check name: %s, section: %s, desc: %s, result: %v", cr.name, cr.section, cr.message, cr.res)
}

type doCheck func(chk Check, las *Las) CheckRes

// Check - отдельная проверка, обязан реализовать функцию doCheck()
type Check struct {
	name    string
	section string
	message string
	do      doCheck
}

// checkResults - слайс с результатами всех проверок
type checkResults []CheckRes

func (crs checkResults) nullWrong() bool {
	for _, r := range crs {
		if r.name == "NullNot0" {
			return true
		}
	}
	return false
}

func (crs checkResults) stepWrong() bool {
	for _, r := range crs {
		if r.name == "StepNot0" {
			return true
		}
	}
	return false
}

func (crs checkResults) wrapWrong() bool {
	for _, r := range crs {
		if r.name == "WrapIsOn" {
			return true
		}
	}
	return false
}

func (crs checkResults) curvesWrong() bool {
	for _, r := range crs {
		if r.name == "CurvesNotPresent" {
			return true
		}
	}
	return false
}

func (crs checkResults) strtStopWrong() bool {
	for _, r := range crs {
		if r.name == "StrtStop" {
			return true
		}
	}
	return false
}

// Checker - ПРОВЕРЩИК, содержит в себе всех отдельных проверщиков,
// методом check() вызавает последовательно всех своих проверщиков,
// результаты отправляет в Logger
type Checker []Check

func (c Checker) check(las *Las) checkResults {
	res := make([]CheckRes, 0)
	for _, chk := range c {
		r := chk.do(chk, las)
		if !r.res {
			res = append(res, r)
		}
	}
	return res
}

/****************************/

// NewEmptyChecker - создание нового примитивного объекта ПРОВЕРЩИКА
// проверяет только на не пустоту объекта las
func NewEmptyChecker() Checker {
	return Checker{
		{chkEmptyName, chkEmptySection, "", emptyCheck},
	}
}

const (
	chkEmptyName    = "LasNotNil"
	chkEmptySection = "NoN"
)

// simpleCheck - return true if las not empty
func emptyCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, nil, !las.IsEmpty()}
}

/****************************/

// NewWrongChecker - создание нового ПРОВЕРЩИКА на ошибочность las файла.
// WRAP = ON
// section ~Curve is empty
func NewWrongChecker() Checker {
	return Checker{
		newWrapCheck(),
		newNotPresentCurvesCheck(),
	}
}

func newWrapCheck() Check {
	return Check{"WrapIsOn", "~V", "WRAP = ON", wrapOn}
}

func wrapOn(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, fmt.Errorf("Wrapped files not support"), !las.IsWraped()}
}

func newNotPresentCurvesCheck() Check {
	return Check{"CurvesNotPresent", "~C", "Curve section is eptry", curvesIsEmpty}
}

func curvesIsEmpty(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, fmt.Errorf("Curve section not exist"), len(las.Logs) > 0}
}

// NewStdChecker - создание нового стандартного ПРОВЕРЩИКА
// проверяет las
// STEP == 0
// NULL == 0
// STRT == STOP
// WELL is empty
func NewStdChecker() Checker {
	return Checker{
		newStepCheck(),
		newNullCheck(),
		newStrtStopCheck(),
		newWellIsEmptyCheck(),
	}
}

// Step Check
func newStepCheck() Check {
	return Check{"StepNot0", "~W", "STEP == 0", stepCheck}
}

func stepCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, nil, las.Step != 0.0}
}

// Null Check
func newNullCheck() Check {
	return Check{"NullNot0", "~W", "NULL == 0", nullCheck}
}

func nullCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, nil, las.Null != 0.0}
}

// STRT == STOP Check
func newStrtStopCheck() Check {
	return Check{"StrtStop", "~W", "STRT == STOP", strtStopCheck}
}

func strtStopCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, nil, las.Strt != las.Stop}
}

// WELL == "" Check
func newWellIsEmptyCheck() Check {
	return Check{"WellNotEmpty", "~W", "WELL == ''", wellIsEmptyCheck}
}

func wellIsEmptyCheck(chk Check, las *Las) CheckRes {
	return CheckRes{chk.name, chk.section, chk.message, nil, len(las.Well) != 0}
}
