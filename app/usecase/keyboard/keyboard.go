package keyboard

import (
	"math"
)

type coordinate struct {
	x, y int
}

type Keyboard struct {
	coordinates []coordinate
}

func NewQWERTY() (*Keyboard, error) {
	return New(QWERTY())
}

func New(layout Layout) (*Keyboard, error) {
	const maxChar = ^byte(0)
	coordinates := make([]coordinate, maxChar)

	for i := 0; i < len(layout); i++ {
		for j := 0; j < len(layout[i]); j++ {
			char := layout[i][j]

			idx := getIdx(char)

			coordinates[idx] = coordinate{i, j}
		}
	}

	return &Keyboard{
		coordinates: coordinates,
	}, nil
}

func (k *Keyboard) GetDistance(a, b byte) (int, error) {
	aIdx := getIdx(a)

	bIdx := getIdx(b)

	aCoord := &k.coordinates[aIdx]
	bCoord := &k.coordinates[bIdx]

	// A bit overhead for int -> float64 -> int conversions, may be fixed by custom Abs function
	return int(math.Abs(float64(aCoord.x-bCoord.x)) + math.Abs(float64(aCoord.y-bCoord.y))), nil
}

func getIdx(char byte) int {
	return int(char)
}
