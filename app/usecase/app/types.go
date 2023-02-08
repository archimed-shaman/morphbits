package app

const (
	uniqueWords    = true // Use only unique words in the password
	bestWordsCount = 12   // Words with the shortest distance for each word length
	passWords      = 4    // Number of words in the password
	minPassLength  = 20   // Minimum password length
	maxPassLength  = 24   // Maximum password length
)

type wordLenMap map[int][]wItem

type DictReader interface {
	Run(handler func(word string) error) error
}

type DistanceCalculator interface {
	GetDistance(a, b byte) (int, error)
}

type Metrics interface {
	IncWords()
	IncFilteredWords()
}

// wItem store the word itself and it's internal distance.
type wItem struct {
	Data string
	Dist int
}
