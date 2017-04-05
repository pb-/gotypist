package main

import (
	"time"

	"unicode/utf8"
)

const speedErrorRatio = 0.2

func speedScore(text string, time time.Duration) float64 {
	return 10 * float64(utf8.RuneCountInString(text)) / (1 + time.Seconds())
}

func errorScore(text string, errors int) float64 {
	return 10 * float64(utf8.RuneCountInString(text)) / (1 + float64(errors))
}

func score(text string, time time.Duration, errors int) float64 {
	return speedErrorRatio*speedScore(text, time) +
		(1-speedErrorRatio)*errorScore(text, errors)
}

func weightedScore(fast, slow, normal float64) float64 {
	return 0.15*fast + 0.35*slow + 0.5*normal
}
