package main

type Finger int

var FingerMap map[rune]Finger

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

var sourceMap = map[Finger]string{
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

func init() {
	FingerMap = make(map[rune]Finger)

	for finger, runes := range sourceMap {
		for _, r := range runes {
			if _, ok := FingerMap[r]; !ok {
				FingerMap[r] = NoFinger
			}
			FingerMap[r] |= finger
		}
	}
}
