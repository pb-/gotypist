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
	scoreCharFactor = 10.
)

// Score between 0 and 1, 0.5 at 2 CPS
func speedScore(text string, time time.Duration) float64 {
	spc := time.Seconds() / float64(utf8.RuneCountInString(text))
	return 1. / (1. + 2*spc)
}

// Score between 0 and 1, 0.5 at 1 error
func errorScore(text string, errors int) float64 {
	return 1. / (1. + float64(errors))
}

func score(text string, time time.Duration, errors int) float64 {
	return speedErrorRatio*speedScore(text, time) +
		(1-speedErrorRatio)*errorScore(text, errors)
}

func finalScore(text string, fast, slow, normal float64) float64 {
	return maxScore(text) * math.Pow(0.15*fast+0.35*slow+0.5*normal, 2)
}

func maxScore(text string) float64 {
	return scoreCharFactor * float64(utf8.RuneCountInString(text))
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
