package app

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	pkgerr "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"morphbits.io/app/usecase/combin"
	"morphbits.io/app/usecase/utils"
)

type App struct {
	dictReader DictReader
	calc       DistanceCalculator
	metrics    Metrics

	words wordLenMap
}

func New(metrics Metrics, dictReader DictReader, calc DistanceCalculator) *App {
	return &App{
		dictReader: dictReader,
		calc:       calc,
		metrics:    metrics,

		words: make(wordLenMap),
	}
}

func (app *App) Run() error {
	if err := app.dictReader.Run(app.handleWord); err != nil {
		return pkgerr.Wrap(err, "failed read dictionary")
	}

	distDict := make([]int, 0, len(app.words))
	for i := range app.words {
		distDict = append(distDict, i)
	}

	bestPass := getBestPass(distDict, app.words, app.calc)

	for i := 0; i < len(bestPass); i++ {
		log.WithFields(log.Fields{
			"pass": bestPass[i].Data,
			"dist": bestPass[i].Dist,
		}).Info("Best pass")
	}

	return nil
}

func (app *App) handleWord(rawWord string) error {
	word := strings.ToLower(rawWord)

	dist, err := calcInternalDistance(word, app.calc)
	if err != nil {
		return err
	}

	length := len(word)

	app.metrics.IncWords()

	shortestWords := app.words[length]
	i := sort.Search(len(shortestWords), func(i int) bool { return shortestWords[i].Dist >= dist })

	if i < len(shortestWords) {
		// Filter words with similar length, similar internal distance
		// and similar start & stop. They have the same
		// distance between the neightbour words.
		foundWord := shortestWords[i].Data
		if shortestWords[i].Dist == dist &&
			foundWord[0] == rawWord[0] &&
			foundWord[len(foundWord)-1] == rawWord[len(rawWord)-1] {
			return nil
		}
	}

	shortestWords = utils.Insert(shortestWords, wItem{
		Data: word,
		Dist: dist,
	}, i)

	end := len(shortestWords)
	if end > bestWordsCount {
		end = bestWordsCount
	} else {
		app.metrics.IncFilteredWords()
	}

	shortestWords = shortestWords[:end]
	app.words[length] = shortestWords

	return nil
}

// getBestPass looks for the best word sequences in the each group of words.
func getBestPass(distDict []int, words wordLenMap, calc DistanceCalculator) []wItem {
	lenCombinations := getLenCombinations(distDict, passWords)
	bestDist := utils.MaxInt()

	var (
		bestPass []wItem
		wg       sync.WaitGroup
		lock     sync.Mutex
	)

	for i := 0; i < len(lenCombinations); i++ {
		lenComb := (*[passWords]int)(lenCombinations[i])

		wg.Add(1)

		go func() {
			defer wg.Done()

			pass, err := getBestPassInGroup(getWords(words, *lenComb), calc, uniqueWords)
			if err != nil {
				log.Fatal(err)
			}

			lock.Lock()
			defer lock.Unlock()

			if pass.Dist == bestDist {
				bestPass = append(bestPass, *pass)
			}

			if pass.Dist < bestDist {
				bestDist = pass.Dist
				bestPass = []wItem{*pass}
			}
		}()
	}

	wg.Wait()

	return bestPass
}

// getBestPassInGroup looks for the best combination within the group of words.
func getBestPassInGroup(words *[passWords][]wItem, calc DistanceCalculator, unique bool) (*wItem, error) {
	var password *wItem

	bestDist := utils.MaxInt()
	dict := utils.MakeRange(0, bestWordsCount-1)

	idxGen := mkIdxGen(dict, words, unique)

	// Iterate over different combinations of the best words to find the best pass
	for idxGen.Next() {
		idx := idxGen.Combination(nil)

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

		dist := dist01 + dist12 + dist23 + word0.Dist + word1.Dist + word2.Dist + word3.Dist

		if dist < bestDist {
			bestDist = dist
			password = &wItem{
				Data: fmt.Sprintf("%s %s %s %s", word0.Data, word1.Data, word2.Data, word3.Data),
				Dist: dist,
			}
		}
	}

	return password, nil
}

func mkIdxGen(dict []int, words *[passWords][]wItem, unique bool) *combin.Generator[int] {
	return combin.NewGenerator(dict, passWords, func(idx []int) bool {
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
}
