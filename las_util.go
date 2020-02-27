package glasio

import (
	"bufio"
	"errors"
	"os"

	"github.com/softlandia/cpd"
)

func isIgnoredLine(s string) bool {
	if (len(s) == 0) || (s[0] == '#') {
		return true
	}
	return false
}

// LoadLasHeader - utility function, if need read only header without data
func LoadLasHeader(fileName string) (*Las, error) {
	las := NewLas()
	iFile, err := os.Open(fileName)
	if err != nil {
		return nil, errors.New("could not open file: '" + fileName + "'")
	}
	defer iFile.Close()
	las.File = iFile
	las.FileName = fileName
	las.Reader, err = cpd.NewReader(las.File)
	las.scanner = bufio.NewScanner(las.Reader)
	if err != nil {
		return nil, err
	}
	las.LoadHeader()
	return las, nil
}
