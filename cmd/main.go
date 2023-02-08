package main

import (
	"os"
	"runtime/pprof"
	"time"

	log "github.com/sirupsen/logrus"
	"morphbits.io/app/interface/dictionary"
	"morphbits.io/app/interface/metrics"
	"morphbits.io/app/usecase/app"
	"morphbits.io/app/usecase/keyboard"
)

func main() {
	start := time.Now()
	////////////////////////////////////
	//// go tool pprof main main.prof
	f, err := os.Create("main.prof")
	if err != nil {
		log.Fatal(err.Error())
	}

	if err = pprof.StartCPUProfile(f); err != nil {
		log.WithField("err", err).Warn("Failed run profiler")
	}

	defer pprof.StopCPUProfile()
	////////////////////////////////////

	log.SetFormatter(&log.TextFormatter{ //nolint:exhaustruct // other fields are defaults
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})

	m := metrics.New()

	kbd, err := keyboard.NewQWERTY()
	if err != nil {
		log.WithField("err", err).Info("Failed init keyboard")
		return
	}

	englishWords := os.Getenv("DICT")
	if englishWords == "" {
		englishWords = "./data/corncob_lowercase.txt"
	}

	dictReader := dictionary.NewFileReader(englishWords)

	application := app.New(m, dictReader, kbd)

	if err := application.Run(); err != nil {
		log.WithField("err", err).Info("Application terminated with error code")
		return
	}

	log.WithField("elapsed", time.Since(start)).Info("Done")
	log.WithFields(m.GetMetrics()).Info("Metrics")
}
