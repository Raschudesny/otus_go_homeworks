package hw02_unpack_string //nolint:golint,stylecheck

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	var sb strings.Builder
	inputRunes := []rune(input)
	for i := 0; i < len(inputRunes); i++ {
		currentRune := inputRunes[i]
		prevRune := currentRune
		if unicode.IsDigit(currentRune) {
			return "", ErrInvalidString
		}
		if i != len(inputRunes)-1 {
			currentRune = inputRunes[i+1]
			if unicode.IsDigit(currentRune) {
				times, err := strconv.Atoi(string(currentRune))
				if err != nil {
					return "", err
				}
				sb.WriteString(strings.Repeat(string(prevRune), times))
				i++
				continue
			}
		}
		sb.WriteRune(prevRune)
	}

	return sb.String(), nil
}
