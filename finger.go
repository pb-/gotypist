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
	LeftPinky:   "1qazQAZ`~!",
	LeftRing:    "2wsxWSX@",
	LeftMiddle:  "3edcEDC#",
	LeftIndex:   "4rfvRFV5tgbTGB$%",
	LeftThumb:   " ",
	RightThumb:  " ",
	RightIndex:  "6yhnYHN7ujmUJM^&",
	RightMiddle: "8ikIK*,<",
	RightRing:   "9olOL(.>",
	RightPinky:  "0pP)-_=+[{]}\\|'\"/?",
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
