package utils

import (
	"os"
	"path/filepath"
)

func OpenFile(fileName string) (*os.File, func() error, error) {
	fn, err := filepath.Abs(fileName)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	file, err := os.Open(fn)

	closer := func() error {
		err := file.Close()
		return err
	}

	return file, closer, err
}
