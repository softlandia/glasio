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

//============================================================================
func main() {

	test1()
	test2()
}

// write empty las file
func test2() {
	las := glasio.NewLas(cpd.CP866)
	las.Null = -99.99
	las.Rkb = 100.01
	las.Strt = 0.201
	las.Stop = 10.01
	las.Step = 0.01
	las.Well = "Примерная-101/бис"
	curve := glasio.NewLasCurve("SP.mV :spontaniously")
	las.Logs["SP"] = curve
	curve.Init(0, "SP", "SP", 5)
	err := las.Save("empty.las")
	log.Printf("err: %v", err)
}

// read and write las files
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
