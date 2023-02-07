package usecase

import (
	"fmt"
	"strings"
	"sync"

	pkgerr "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"
	"morphbits.io/app/usecase/combin"
)

const (
	passWords     = 4
	minPassLength = 20
	maxPassLength = 24
)

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

type wItem struct {
	Data         string
	InternalDist int
}

type WordGroup map[string]*wItem

type App struct {
	dictReader DictReader
	calc       DistanceCalculator
	metrics    Metrics

	filteredWords map[int]WordGroup
}

func NewApp(metrics Metrics, dictReader DictReader, calc DistanceCalculator) *App {
	return &App{
		dictReader: dictReader,
		calc:       calc,
		metrics:    metrics,

		filteredWords: make(map[int]WordGroup),
	}
}

func (app *App) Run() error {
	if err := app.dictReader.Run(app.handleWord); err != nil {
		return pkgerr.Wrap(err, "failed read dictionary")
	}

	for key, value := range app.filteredWords {
		fmt.Println(key, len(value))
	}

	distDict := make([]int, 0, len(app.filteredWords))
	for i := range app.filteredWords {
		distDict = append(distDict, i)
	}

	// Now let's find all possible words' length combinations
	// to get from 20 to 24 password length in sum.

	gen := combin.NewGenerator(distDict, passWords, func(l []int) bool {
		s := sum(l...)
		return s >= minPassLength && s <= maxPassLength
	})

	combinations := make([][]int, 0, 2048)
	for gen.Next() {
		combinations = append(combinations, gen.Combination(nil))
	}

	log.WithField("num", len(combinations)).Info("Possible word length combinations")

	// Now let's find the best word sequences in the each group

	var wg sync.WaitGroup

	for i := 0; i < len(combinations); i++ {
		wordLengthCombination := combinations[i]
		wg.Add(1)

		go func() {
			defer wg.Done()

			pass, err := getBestWords(
				[4]WordGroup{
					app.filteredWords[wordLengthCombination[0]],
					app.filteredWords[wordLengthCombination[1]],
					app.filteredWords[wordLengthCombination[2]],
					app.filteredWords[wordLengthCombination[3]],
				},
				app.calc,
			)
			if err != nil {
				panic(err)
			}

			fmt.Println(pass)
		}()

		// for j := 0; j < len(passwords); j++ {
		// 	fmt.Println(passwords[j])
		// }
	}

	wg.Wait()

	return nil
}

func getBestWords(words [4]WordGroup, calc DistanceCalculator) (*wItem, error) {
	bestDist := int(^uint(0) >> 1)

	var password *wItem

	for _, word0 := range words[0] {
		for _, word1 := range words[1] {
			for _, word2 := range words[2] {
				for _, word3 := range words[3] {
					dist01, err := calcWordDistance(word0.Data, word1.Data, calc)
					if err != nil {
						return nil, err
					}

					dist12, err := calcWordDistance(word1.Data, word2.Data, calc)
					if err != nil {
						return nil, err
					}

					dist23, err := calcWordDistance(word2.Data, word3.Data, calc)
					if err != nil {
						return nil, err
					}

					dist := dist01 + dist12 + dist23 +
						word0.InternalDist + word1.InternalDist + word2.InternalDist + word3.InternalDist

					if dist < int(bestDist) {
						bestDist = dist
						password = &wItem{
							Data:         word0.Data + word1.Data + word2.Data + word3.Data,
							InternalDist: dist,
						}
					}
				}
			}
		}
	}

	return password, nil
}

func (app *App) handleWord(rawWord string) error {
	word := strings.ToLower(rawWord)

	dist, err := calcInternalDistance(word, app.calc)
	if err != nil {
		return err
	}

	length := len(word)
	key := mkKey(word)

	app.metrics.IncWords()

	if m, found := app.filteredWords[length]; !found {
		app.filteredWords[length] = WordGroup{
			key: {
				Data:         word,
				InternalDist: dist,
			},
		}

		app.metrics.IncFilteredWords()
	} else {
		if otherWord, found := m[key]; !found {
			m[key] = &wItem{
				Data:         word,
				InternalDist: dist,
			}

			app.metrics.IncFilteredWords()
		} else if otherWord.InternalDist > dist {
			// Just override current fields to avoid reallocations
			otherWord.Data = word
			otherWord.InternalDist = dist
		}
	}

	return nil
}

func calcInternalDistance(word string, calc DistanceCalculator) (int, error) {
	const minWordSize = 2

	if len(word) < minWordSize {
		return 0, nil
	}

	distance := 0

	for i := 1; i < len(word); i++ {
		a := word[i]
		b := word[i-1]

		d, err := calc.GetDistance(a, b)
		if err != nil {
			return 0, pkgerr.Wrapf(err, "error occurred while calculating distance for word '%s'", word)
		}

		distance += d
	}

	return distance, nil
}

func calcWordDistance(word1, word2 string, calc DistanceCalculator) (int, error) {
	a := word1[len(word1)-1]
	b := word2[0]

	distance, err := calc.GetDistance(a, b)
	if err != nil {
		return 0, pkgerr.Wrapf(err, "error occurred while calculating distance for words '%s', '%s'",
			word1, word2)
	}

	return distance, nil
}

func mkKey(word string) string {
	if word == "" {
		return ""
	}

	var sb strings.Builder

	sb.WriteByte(word[0])
	sb.WriteByte(word[len(word)-1])

	return sb.String()
}

func sum[A constraints.Ordered](values ...A) A {
	var sum A

	for i := 0; i < len(values); i++ {
		sum += values[i]
	}

	return sum
}
