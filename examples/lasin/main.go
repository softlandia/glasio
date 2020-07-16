package main

import (
	"fmt"
	"os"

	"github.com/softlandia/glasio"
)

// Sample
// read one file and print all warnings to stdout:
// warning number, number of line in las file, message
func main() {
	if len(os.Args) == 1 {
		fmt.Printf("using:\nlasin fileName.las\n")
		os.Exit(0)
	}
	glasio.MaxWarningCount = 100
	las := glasio.NewLas()
	_, err := las.Open(os.Args[1])
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("%s", las.Warnings.ToString())
}
