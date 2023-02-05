package main

import (
	log "github.com/sirupsen/logrus"
	"morphbits.io/app/interface/dictionary"
	"morphbits.io/app/usecase"
	"morphbits.io/app/usecase/keyboard"
)

func main() {
	kbd, err := keyboard.NewQWERTY()
	if err != nil {
		log.WithField("err", err).Info("Failed init keyboard")
		return
	}

	dictReader := dictionary.NewFileReader("./data/corncob_lowercase.txt")

	app := usecase.NewApp(dictReader, kbd)

	if err := app.Run(); err != nil {
		log.WithField("err", err).Info("Application terminated with error code")
		return
	}
}
