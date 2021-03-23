package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"unicode"
)

var ErrEmptyDirName = errors.New("directory name is empty")

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

const EscapedTerminalNullBytes = 0x00

// ReadFirstLineFromFile reads one line from a specified file or fails with error.
func ReadFirstLineFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading %s file: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Println(closeErr)
		}
	}()

	bufReader := bufio.NewReader(file)
	readString, err := bufReader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("error reading first line from %s file: %w", filePath, err)
	}
	return readString, nil
}

// ReadDirEnvs reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDirEnvs(dir string) (Environment, error) {
	if len(dir) == 0 {
		return nil, ErrEmptyDirName
	}

	dirFileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading %s directory files: %w", dir, err)
	}
	if len(dirFileInfos) == 0 {
		return nil, nil
	}

	envs := make(Environment, len(dirFileInfos))
	for _, envFileInfo := range dirFileInfos {
		if !strings.Contains(envFileInfo.Name(), "=") && envFileInfo.Mode().IsRegular() {
			if envFileInfo.Size() == 0 {
				envs[envFileInfo.Name()] = EnvValue{"", true}
				continue
			}

			line, err := ReadFirstLineFromFile(path.Join(dir, envFileInfo.Name()))
			if err != nil {
				log.Println(err)
				continue
			}
			line = string(bytes.ReplaceAll([]byte(line), []byte{EscapedTerminalNullBytes}, []byte("\n")))
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			envs[envFileInfo.Name()] = EnvValue{line, len(line) == 0}
		}
	}
	return envs, nil
}
