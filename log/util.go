package log

import (
	"os"
)

func syncFile(pathname string) (err error) {

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

func syncDirectory(path string) (err error) {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}
