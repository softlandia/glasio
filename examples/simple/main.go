// glas
// Copyright 2018 softlandia@gmail.com
// Обработка las файлов. Построение словаря и замена мнемоник на справочные

package main

import (
	"fmt"
	"log"

	"github.com/softlandia/cpd"

	"github.com/softlandia/glasio"
)

func main() {

	test1()
	test2()
}

// reads las file 
// writes messages file 
// and writes the repaired las files
func test1() {
	//test file "1.las"
	las := glasio.NewLas()
	n, err := las.Open("expand_points_01.las")
	if n != 7 {
		fmt.Printf("TEST read expand_points_01.las ERROR, n = %d, must 7\n", n)
		fmt.Println(err)
	}
	las.SaveWarning("expand_points_01.warning.md")

	err = las.Save("expand_points_01+.las")
	if err != nil {
		fmt.Println("TEST save -1.las ERROR: ", err)
	}

	las = nil
}

// writes simple las file
func test2() {
	las := glasio.NewLas(cpd.CP866)
	las.Null = -99.99
	las.Rkb = 100.01
	las.Strt = 0.201
	las.Stop = 10.01
	las.Step = 0.01
	las.Well = "Примерная-101/бис"
	d := glasio.NewLasCurve("DEPT")
	d.Init(0, "DEPT", "DEPT", 5)
	sp := glasio.NewLasCurve("SP.mV :spontaniously")
	sp.Init(1, "SP", "SP", 5)
	for i := 0; i < 5; i++ {
		d.D[i] = float64(i)
		d.V[i] = float64(i) / 100
		sp.D[i] = float64(i)
		sp.V[i] = float64(i) / 100
	}
	las.Logs["DEPT"] = d
	las.Logs["SP"] = sp
	las.SetActuallyNumberPoints(5)
	err := las.Save("empty.las")
	log.Printf("err: %v", err)
}
