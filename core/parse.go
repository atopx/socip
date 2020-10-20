package core

import (
	"bufio"
	"os"
	"strings"
)

func LoadSource(filename string, names chan string, reader chan error) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, `"minip"`) {
			names <- line
		}
	}

	close(names)
	reader <- scanner.Err()
}

func ParseLine(line string) (fields []string) {
	line = strings.ReplaceAll(line, `"`, "")
	fields = strings.Split(line, "\t")
	return fields
}
