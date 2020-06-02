package main

import (
	"fmt"
	"os"

	"github.com/softlandia/glasio"
)

//Пример чтения и сохранения нескольких LAS файлов
func main() {
	if len(os.Args) == 1 {
		fmt.Printf("using:\nlasin fileName.las\n")
		os.Exit(0)
	}
	las := glasio.NewLas()
	_, err := las.Open(os.Args[1])
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("%s", las.Warnings.ToString())
}
