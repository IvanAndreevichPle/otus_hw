package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	var result strings.Builder
	var previousChar rune
	var isShielded bool
	var isPreviousDigit bool

	for _, char := range s {
		if unicode.IsDigit(char) && !isShielded {
			if previousChar == 0 || isPreviousDigit {
				return "", ErrInvalidString
			}
			count, err := strconv.Atoi(string(char))
			if err != nil {
				return "", err
			}
			handleDigit(&result, previousChar, count)
			isPreviousDigit = true
		} else {
			handleChar(&result, char, &isShielded)
			previousChar = char
			isPreviousDigit = false
		}
	}

	if isShielded {
		return "", ErrInvalidString
	}

	return result.String(), nil
}

func handleDigit(result *strings.Builder, previousChar rune, count int) {
	previousCharStr := string(previousChar)
	if count == 0 {
		str := result.String()
		result.Reset()
		result.WriteString(str[:len(str)-len(previousCharStr)])
	} else {
		result.WriteString(strings.Repeat(previousCharStr, count-1))
	}
}

func handleChar(result *strings.Builder, char rune, isShielded *bool) {
	charStr := string(char)
	switch {
	case *isShielded:
		result.WriteString(charStr)
		*isShielded = false
	case char == '\\':
		*isShielded = true
	default:
		result.WriteString(charStr)
	}
}
