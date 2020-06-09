package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/softlandia/glasio"
)

//Пример чтения и сохранения нескольких LAS файлов
func main() {
	fileList := make([]string, 0, 10)
	n := findFilesExt(&fileList, ".", ".las")
	fmt.Printf("files found: %d\n", n)
	if n == 0 {
		os.Exit(0)
	}
	for _, f := range fileList {
		repaireLas(f)
	}
}

func repaireLas(fileName string) {
	las := glasio.NewLas()
	las.FileName = fileName
	fmt.Printf("file: '%s'", fileName)
	n, err := las.Open(fileName) // считываем файл
	fmt.Printf(" read\n")
	// примеры проверки прочитанного las файла
	if las.IsWraped() {
		fmt.Printf("wrapped\n")
		return
	}
	if n == 0 {
		fmt.Printf("data not exist\n")
		return
	}
	if err != nil {
		fmt.Printf("error: %v\n", err) // Open() неохотно возвращает err, скорее всего ну вообще не получается прочитать
	}
	las.Save(las.FileName + "-") //сохраняем с символом минус в расширении
}

func findFilesExt(fileList *[]string, path, fileNameExt string) int {
	extFile := strings.ToUpper(fileNameExt)
	index := 0 //index founded files
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil //skip folders
		}
		if strings.ToUpper(filepath.Ext(path)) != extFile {
			return nil //skip files with wrong extention
		}
		index++
		*fileList = append(*fileList, path)
		return nil
	})
	return index
}
