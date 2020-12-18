package tcp

import (
	"errors"
)

var (
	ErrUnknownError = errors.New("tcp: unknown error")

	defaultErrorCode    = 0
	defaultErrorMessage = ErrUnknownError

	errorsCodes = map[error]int{
		// TODO: add tcp error codes.
	}

	errorsMessages = map[int]error{
		// TODO: add tcp error messages.
	}
)

func GetErrorCode(err error) (code int) {

	code, ok := errorsCodes[err]
	if !ok {
		return defaultErrorCode
	}

	return code
}

func GetErrorMessage(code int) (err error) {

	err, ok := errorsMessages[code]
	if !ok {
		return defaultErrorMessage
	}

	return err
}
