package main

type Finger int

const (
	NoFinger Finger = 0

	LeftPinky Finger = 1 << iota
	LeftRing
	LeftMiddle
	LeftIndex
	LeftThumb

	RightThumb
	RightIndex
	RightMiddle
	RightRing
	RightPinky
)

var FingerSequence = []Finger{
	LeftPinky, LeftRing, LeftMiddle, LeftIndex, LeftThumb,
	RightThumb, RightIndex, RightMiddle, RightRing, RightPinky,
}

var preMap = map[Finger]string{
	LeftPinky:   "1qaz",
	LeftRing:    "2wsx",
	LeftMiddle:  "3edc",
	LeftIndex:   "4rfv5tgb",
	LeftThumb:   " ",
	RightThumb:  " ",
	RightIndex:  "6yhn7ujm",
	RightMiddle: "8ik",
	RightRing:   "9ol",
	RightPinky:  "0p",
}

var FingerMap = compileFingerMap()

func compileFingerMap() map[rune]Finger {
	m := make(map[rune]Finger)

	for finger, runes := range preMap {
		for _, r := range runes {
			if _, ok := m[r]; !ok {
				m[r] = NoFinger
			}
			m[r] |= finger
		}
	}

	return m
}
