package main

import (
	"math"
	"time"

	"unicode/utf8"
)

const (
	speedErrorRatio = 0.2
	scoreScalar     = 100.
	scoreExponent   = 2.3
)

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

func requiredScore(level int) float64 {
	return scoreScalar * math.Pow(float64(level), scoreExponent)
}

func level(score float64) int {
	return int(math.Pow(score/scoreScalar, 1./scoreExponent))
}

func progress(score float64) float64 {
	currentLevel := level(score)
	currentLevelScore := requiredScore(currentLevel)
	nextLevelScore := requiredScore(currentLevel + 1)
	return (score - currentLevelScore) / (nextLevelScore - currentLevelScore)
}
