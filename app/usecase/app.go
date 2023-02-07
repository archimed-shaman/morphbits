package usecase

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	pkgerr "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"
	"morphbits.io/app/usecase/combin"
)

const (
	bestWordsCount = 12
	passWords      = 4
	minPassLength  = 20
	maxPassLength  = 24
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

type App struct {
	dictReader DictReader
	calc       DistanceCalculator
	metrics    Metrics

	filteredWords map[int][]wItem
}

func NewApp(metrics Metrics, dictReader DictReader, calc DistanceCalculator) *App {
	return &App{
		dictReader: dictReader,
		calc:       calc,
		metrics:    metrics,

		filteredWords: make(map[int][]wItem),
	}
}

func (app *App) Run() error {
	if err := app.dictReader.Run(app.handleWord); err != nil {
		return pkgerr.Wrap(err, "failed read dictionary")
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

	bestDist := int(^uint(0) >> 1)
	var bestPass []wItem

	var wg sync.WaitGroup
	var lock sync.Mutex

	for i := 0; i < len(combinations); i++ {
		wordLengthCombination := combinations[i]

		wg.Add(1)

		go func() {
			defer wg.Done()

			pass, err := getBestWords(
				[4][]wItem{
					app.filteredWords[wordLengthCombination[0]],
					app.filteredWords[wordLengthCombination[1]],
					app.filteredWords[wordLengthCombination[2]],
					app.filteredWords[wordLengthCombination[3]],
				},
				app.calc,
				true,
			)
			if err != nil {
				panic(err)
			}

			lock.Lock()
			defer lock.Unlock()

			if pass.InternalDist == bestDist {
				bestPass = append(bestPass, *pass)
			}

			if pass.InternalDist < bestDist {
				bestDist = pass.InternalDist
				bestPass = []wItem{*pass}
			}
		}()
	}

	wg.Wait()

	for i := 0; i < len(bestPass); i++ {
		log.WithFields(log.Fields{
			"pass": bestPass[i].Data,
			"dist": bestPass[i].InternalDist,
		}).Info("Best pass")
		// fmt.Println(bestPass[i])
	}

	return nil
}

func getBestWords(words [passWords][]wItem, calc DistanceCalculator, unique bool) (*wItem, error) {
	bestDist := int(^uint(0) >> 1)

	var password *wItem

	dict := makeRange(0, bestWordsCount-1)

	gen := combin.NewGenerator(dict, passWords, func(idx []int) bool {
		if len(idx) != len(words) {
			panic("bad indexes")
		}

		keys := make(map[int]bool)

		for i := 0; i < len(idx); i++ {
			if idx[i] >= len(words[0]) {
				return false
			}

			keys[idx[i]] = true
		}

		if unique && len(keys) != len(idx) {
			return false
		}

		return true
	})

	for gen.Next() {
		idx := gen.Combination(nil)
		word0 := words[0][idx[0]]
		word1 := words[1][idx[1]]
		word2 := words[2][idx[2]]
		word3 := words[3][idx[3]]

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
				Data: fmt.Sprintf("%s %s %s %s",
					word0.Data, word1.Data, word2.Data, word3.Data),
				InternalDist: dist,
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

	app.metrics.IncWords()

	shortestWords := app.filteredWords[length]
	i := sort.Search(len(shortestWords), func(i int) bool { return shortestWords[i].InternalDist >= dist })

	if i < len(shortestWords) {
		// Filter words with similar length, similar internal distance
		// and similar start & stop. They have the same
		// distance between the neightbour words.

		foundWord := shortestWords[i].Data
		if shortestWords[i].InternalDist == dist &&
			foundWord[0] == rawWord[0] &&
			foundWord[len(foundWord)-1] == rawWord[len(rawWord)-1] {
			return nil
		}
	}

	shortestWords = insert(shortestWords, wItem{
		Data:         word,
		InternalDist: dist,
	}, i)

	end := len(shortestWords)
	if end > bestWordsCount {
		end = bestWordsCount
	}

	shortestWords = shortestWords[:end]
	app.filteredWords[length] = shortestWords

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

func sum[A constraints.Ordered](values ...A) A {
	var sum A

	for i := 0; i < len(values); i++ {
		sum += values[i]
	}

	return sum
}

func insert[A any](slice []A, value A, position int) []A {
	if position < 0 || position > len(slice) {
		panic("wrong index")
	}

	var a A
	slice = append(slice, a)
	copy(slice[position+1:], slice[position:])
	slice[position] = value

	return slice
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}
