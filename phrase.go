package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// PhraseFunc generates a phrase from a seed
type PhraseFunc func(seed int64) string

// DefaultPhrase generates a built-in fallback phrase
func DefaultPhrase(seed int64) string {
	return "the quick brown fox jumps over the lazy dog"
}

// StaticPhrase returns a static phrase generator function with given phrase
func StaticPhrase(phrase string) PhraseFunc {
	return func(seed int64) string {
		return phrase
	}
}

// RandomPhrase composes a random phrase with given length from given words
func RandomPhrase(words []string, minLength int, numProb float64) PhraseFunc {
	return func(seed int64) string {
		rand := rand.New(rand.NewSource(seed))
		var phrase []string
		l := -1
		for l < minLength {
			var w string
			if rand.Float64() < numProb {
				w = strconv.FormatInt(rand.Int63n(10000), 10)
			} else {
				w = words[rand.Int31n(int32(len(words)))]
			}
			phrase = append(phrase, w)
			l += 1 + len(w)
		}
		return strings.Join(phrase, " ")
	}
}

func loadDictionary(path string) ([]string, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	return filterWords(lines, `^[a-z]+$`, 8), nil
}

func filterWords(words []string, pattern string, minLength int) []string {
	filtered := make([]string, 0)
	compiled := regexp.MustCompile(pattern)

	for _, word := range words {
		trimmed := strings.TrimSpace(word)
		if compiled.MatchString(trimmed) && len(trimmed) < minLength {
			filtered = append(filtered, trimmed)
		}
	}

	return filtered
}

func readLines(path string) ([]string, error) {
	var lines []string

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open line file: %s", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return lines, nil
			}
			return nil, fmt.Errorf("failed to read lines: %s", err)
		}
		lines = append(lines, line[:len(line)-1])
	}

	return lines, nil
}
