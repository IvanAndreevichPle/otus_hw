package hw03frequencyanalysis

import (
	"sort"
	"strings"
	"unicode"
)

const limit = 10

func Top10(input string) []string {
	if input == "" {
		return nil
	}

	words := strings.Fields(input)
	wordsMap := make(map[string]int)

	for _, word := range words {
		cleaned := strings.ToLower(
			strings.TrimFunc(word, func(r rune) bool {
				return r != '-' && unicode.IsPunct(r)
			}),
		)

		if cleaned == "" || cleaned == "-" {
			continue
		}

		wordsMap[cleaned]++
	}

	keys := make([]string, 0, len(wordsMap))
	for key := range wordsMap {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if wordsMap[keys[i]] == wordsMap[keys[j]] {
			return keys[i] < keys[j]
		}
		return wordsMap[keys[i]] > wordsMap[keys[j]]
	})

	return keys[:min(len(keys), limit)]
}
