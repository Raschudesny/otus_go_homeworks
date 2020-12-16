package hw03_frequency_analysis //nolint:golint,stylecheck
import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type wordFrequency struct {
	word      string
	frequency int64
}

func EscapeWordPunctuation(word string) string {
	return strings.TrimFunc(word, unicode.IsPunct)
}

func Top10(input string) []string {
	if input == "" {
		return nil
	}

	r := regexp.MustCompile(`\s+`)
	words := r.Split(input, -1)
	if len(words) == 0 {
		return nil
	}

	frequenciesMap := map[string]int64{}
	for _, word := range words {
		escapedWord := EscapeWordPunctuation(strings.ToLower(word))
		if len(escapedWord) > 0 {
			frequenciesMap[escapedWord]++
		}
	}

	textStatistic := make([]wordFrequency, 0, len(frequenciesMap))
	for word, freq := range frequenciesMap {
		textStatistic = append(textStatistic, wordFrequency{word, freq})
	}

	sort.Slice(textStatistic, func(idx1, idx2 int) bool {
		return textStatistic[idx1].frequency > textStatistic[idx2].frequency
	})

	result := make([]string, 0, len(textStatistic))
	for idx, word := range textStatistic {
		if idx > 9 {
			break
		}
		result = append(result, word.word)
	}
	return result
}
