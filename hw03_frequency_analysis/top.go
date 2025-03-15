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
		cleaned := cleanWord(word)

		if cleaned == "" {
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

	result := make([]string, min(len(keys), limit))
	copy(result, keys)
	return result[:min(len(result), limit)]
}

func cleanWord(word string) string {
	isPunctuationOnly := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for _, r := range s {
			if !unicode.IsPunct(r) {
				return false
			}
		}
		return len(s) > 1
	}

	trimPunctuation := func(s string) string {
		return strings.TrimFunc(s, unicode.IsPunct)
	}

	if isPunctuationOnly(word) {
		return word
	}

	return strings.ToLower(trimPunctuation(word))
}
