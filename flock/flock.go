package flock

import (
	"errors"
	"os"
	"syscall"
)

var (
	ErrLocked   = errors.New("flock: locked")
	ErrOrphaned = errors.New("flock: orphaned")
)

type Flock struct {
	pathname string
	mode     os.FileMode
	file     *os.File
}

func New(pathname string, mode os.FileMode) (fl *Flock) {

	fl = &Flock{
		pathname: pathname,
		mode:     mode,
		file:     nil,
	}

	return fl
}

func (fl *Flock) Acquire() (err error) {

	f, err := os.OpenFile(fl.pathname, os.O_RDONLY, os.FileMode(0))
	if err == nil {
		// Lock file exists and has been opened successfuly, try to
		// acquire lock.

		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			// Couldn't acquire lock: the file is locked by someone
			// else.

			return ErrLocked
		}

		err = f.Close()
		if err != nil {
			return err
		}

		return ErrOrphaned
	}

	if !os.IsNotExist(err) {
		return err
	}

	// Open lock file and acquire lock.
	f, err = os.OpenFile(fl.pathname, os.O_WRONLY|os.O_CREATE, fl.mode)
	if err != nil {
		return err
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return err
	}

	fl.file = f

	return nil
}

func (fl *Flock) Release() (err error) {

	err = fl.file.Close()
	if err != nil {
		return err
	}

	err = fl.Clear()
	if err != nil {
		return err
	}

	return nil
}

func (fl *Flock) Clear() (err error) {

	err = os.Remove(fl.pathname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return nil
}
