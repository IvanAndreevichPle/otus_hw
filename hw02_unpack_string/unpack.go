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
	var prevChar rune
	var escaped bool
	var prevWasDigit bool

	for _, char := range s {
		if unicode.IsDigit(char) && !escaped {
			if prevChar == 0 || prevWasDigit {
				return "", ErrInvalidString
			}
			count, _ := strconv.Atoi(string(char))
			handleDigit(&result, prevChar, count)
			prevWasDigit = true
		} else {
			handleChar(&result, char, &escaped)
			prevChar = char
			prevWasDigit = false
		}
	}

	if escaped {
		return "", ErrInvalidString
	}

	return result.String(), nil
}

func handleDigit(result *strings.Builder, prevChar rune, count int) {
	if count == 0 {
		str := result.String()
		result.Reset()
		result.WriteString(str[:len(str)-len(string(prevChar))])
	} else {
		result.WriteString(strings.Repeat(string(prevChar), count-1))
	}
}

func handleChar(result *strings.Builder, char rune, escaped *bool) {
	switch {
	case *escaped:
		result.WriteRune(char)
		*escaped = false
	case char == '\\':
		*escaped = true
	default:
		result.WriteRune(char)
	}
}
