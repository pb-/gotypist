package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
)

func generateText(seed int64, pool []string) string {
	rand := rand.New(rand.NewSource(seed))
	var words []string
	l := -1
	for l < 30 {
		w := pool[rand.Int31n(int32(len(pool)))]
		words = append(words, w)
		l += len(w) + 1
	}
	return strings.Join(words, " ")
}

func readWords(path string) ([]string, error) {
	var words []string

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open words: %s", err)
	}
	defer file.Close()

	pattern := regexp.MustCompile(`^[a-z]+$`)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return words, nil
			}
			return nil, fmt.Errorf("failed to read words: %s", err)
		}

		trimmed := strings.TrimSpace(line)

		if isCode {
			words = append(words, trimmed)
		} else if pattern.MatchString(trimmed) && len(trimmed) < 8 {
			words = append(words, trimmed)
		}
	}
}

func getWords(path string) []string {
	words, err := readWords(path)
	if err != nil || len(words) < 1 {
		return []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
	}
	return words
}
