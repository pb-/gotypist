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

func generateText(seed int64, words []string) string {
	rand := rand.New(rand.NewSource(seed))
	var w []string
	l := int(rand.Int31n(4) + 4)
	for len(w) < l {
		w = append(w, words[rand.Int31n(int32(len(words)))])
	}
	return strings.Join(w, " ")
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
		if pattern.MatchString(trimmed) && len(trimmed) < 8 {
			words = append(words, trimmed)
		}
	}
}

func getWords(path string) []string {
	words, err := readWords(path)
	if err != nil || len(words) < 9 {
		return []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
	}
	return words
}
