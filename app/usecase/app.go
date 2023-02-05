package usecase

import (
	"strconv"
	"strings"

	pkgerr "github.com/pkg/errors"
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

	filteredWords map[string]*wItem
}

func NewApp(metrics Metrics, dictReader DictReader, calc DistanceCalculator) *App {
	return &App{
		dictReader: dictReader,
		calc:       calc,
		metrics:    metrics,

		filteredWords: make(map[string]*wItem),
	}
}

func (app *App) Run() error {
	if err := app.dictReader.Run(app.handleWord); err != nil {
		return pkgerr.Wrap(err, "failed read dictionary")
	}

	return nil
}

func (app *App) handleWord(rawWord string) error {
	word := strings.ToLower(rawWord)

	dist, err := app.calcInternalDistance(word)
	if err != nil {
		return err
	}

	key := mkKey(word)

	app.metrics.IncWords()

	if otherWord, found := app.filteredWords[key]; !found {
		app.filteredWords[key] = &wItem{
			Data:         word,
			InternalDist: dist,
		}

		app.metrics.IncFilteredWords()
	} else if otherWord.InternalDist > dist {
		// Just override current fields to avoid reallocations
		otherWord.Data = word
		otherWord.InternalDist = dist
	}

	return nil
}

func (app *App) calcInternalDistance(word string) (int, error) {
	const minWordSize = 2

	if len(word) < minWordSize {
		return 0, nil
	}

	distance := 0

	for i := 1; i < len(word); i++ {
		a := word[i]
		b := word[i-1]

		d, err := app.calc.GetDistance(a, b)
		if err != nil {
			return 0, pkgerr.Wrapf(err, "error occurred while calculating distance for word '%s'", word)
		}

		distance += d
	}

	return distance, nil
}

func mkKey(word string) string {
	if word == "" {
		return ""
	}

	var sb strings.Builder

	sb.WriteByte(word[0])
	sb.Write([]byte(strconv.Itoa(len(word))))
	sb.WriteByte(word[len(word)-1])

	return sb.String()
}
