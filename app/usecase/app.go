package usecase

import (
	"strings"

	pkgerr "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type DictReader interface {
	Run(handler func(word string)) error
}

type DistanceCalculator interface {
	GetDistance(a, b byte) (int, error)
}

type App struct {
	dictReader DictReader
	calc       DistanceCalculator
}

func NewApp(dictReader DictReader, calc DistanceCalculator) *App {
	return &App{
		dictReader: dictReader,
		calc:       calc,
	}
}

func (app *App) Run() error {
	if err := app.dictReader.Run(app.handleWord); err != nil {
		return pkgerr.Wrap(err, "failed read dictionary")
	}

	return nil
}

func (app *App) handleWord(word string) {
	log.Info(strings.ToLower(word))
}
