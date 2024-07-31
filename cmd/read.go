package cmd

import (
	"bufio"
	"os"
)

func ReadFileLines(filename string) (lines []string, err error) {
	var osFile *os.File

	if osFile, err = os.Open(filename); err == nil {
		defer osFile.Close()
	} else {
		return nil, err
	}

	var scanner = bufio.NewScanner(osFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
