// glas
// Copyright 2018 softlandia@gmail.com
// Обработка las файлов. Построение словаря и замена мнемоник на справочные

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/softlandia/cpd"

	"github.com/softlandia/glasio"
)

func main() {

	readAndSave()
	makeAndSave()
}

// reads las file
// writes messages file
// and writes the repaired las files
func readAndSave() {
	//test file "1.las"
	las := glasio.NewLas()
	n, err := las.Open("ex01.las")
	if n != 7 {
		fmt.Printf("TEST read ex01.las ERROR, n = %d, must 7\n", n)
		fmt.Println(err)
	}
	las.SaveWarning("ex01.warning.md")

	err = las.Save("ex01+.las")
	if err != nil {
		fmt.Println("TEST save -1.las ERROR: ", err)
	}

	las = nil
}

var s = `~W
NULL. -99.99:
STRT.m 00.21:
STOP.m 10.01:
STEP.m 00.01:
WELL. Примерная-101 / бис:`

// make las object from string in memory and writes simple las file
func makeAndSave() {
	//make object, the file will be saved in encoding CP866
	las := glasio.NewLas(cpd.CP866)
	//parse string s as las file, only section ~W
	las.Load(strings.NewReader(s))
	//make curve depth
	d := glasio.NewLasCurve("DEPT", las)
	//make curve sp
	sp := glasio.NewLasCurve("SP.mV", las)
	for i := 0; i < 5; i++ {
		d.D = append(d.D, las.STRT()+las.STEP()*float64(i))
		sp.D = append(sp.D, d.D[i])
		sp.V = append(sp.V, float64(i)/100)
	}
	las.Logs = append(las.Logs, d)
	las.Logs = append(las.Logs, sp)
	err := las.Save("simple.las")
	log.Printf("err: %v", err)
}
