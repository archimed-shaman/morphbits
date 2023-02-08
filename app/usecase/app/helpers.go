package app

import (
	pkgerr "github.com/pkg/errors"
	"morphbits.io/app/usecase/combin"
	"morphbits.io/app/usecase/utils"
)

func getLenCombinations(distDict []int, length, minPassLen, maxPassLen int) [][]int {
	gen := combin.NewGenerator(distDict, length, func(l []int) bool {
		s := utils.Sum(l...)
		return s >= minPassLen && s <= maxPassLen
	})

	const preallocated = 1024

	lenCombinations := make([][]int, 0, preallocated)
	for gen.Next() {
		lenCombinations = append(lenCombinations, gen.Combination(nil))
	}

	return lenCombinations
}

func calcInternalDistance(word string, calc DistanceCalculator) (int, error) {
	const minWordSize = 2

	if len(word) < minWordSize {
		return 0, nil
	}

	distance := 0

	for i := 1; i < len(word); i++ {
		a := word[i-1]
		b := word[i]

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

// getWords makes groups of words with the given length.
func getWords(wordsByLen wordLenMap, lenIdx [passWords]int) *[passWords][]wItem {
	var words [passWords][]wItem

	for i := 0; i < len(lenIdx); i++ {
		words[i] = wordsByLen[lenIdx[i]]
	}

	return &words
}
