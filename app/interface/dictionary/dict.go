package dictionary

import (
	"bufio"
	"os"

	pkgerr "github.com/pkg/errors"
)

type FileReader struct {
	fileName string
}

func NewFileReader(fileName string) *FileReader {
	return &FileReader{
		fileName: fileName,
	}
}

func (fr *FileReader) Run(handler func(word string) error) error {
	f, err := os.Open(fr.fileName)
	if err != nil {
		return pkgerr.Wrapf(err, "failed open file '%s'", fr.fileName)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := handler(scanner.Text()); err != nil {
			return pkgerr.Wrapf(err, "scaning file '%s' aborted due to error", fr.fileName)
		}
	}

	if err := scanner.Err(); err != nil {
		return pkgerr.Wrapf(err, "failed read file '%s'", fr.fileName)
	}

	return nil
}
