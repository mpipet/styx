package api

import (
	"fmt"
)

var (
	defaultErrorCode          = "unknown_error"
	methodNotAllowedErrorCode = "method_not_allowed"
	notFoundErrorCode         = "not_found"
	paramsErrorCode           = "invalid_params"
	logExistErrorCode         = "log_exist"
	logNotFoundErrorCode      = "log_not_found"
	logNotAvailableErrorCode  = "log_not_available"
	missingLengthErrorCode    = "missing_content_length"

	defaultErrorMessage          = "api: unknown error"
	methodNotAllowedErrorMessage = "api: method not allowed"
	notFoundErrorMessage         = "api: not found"
	logExistErrorMessage         = "api: log already exists"
	logNotFoundErrorMessage      = "api: log not found"
	logNotAvailableErrorMessage  = "api: log not available"
	missingLengthErrorMessage    = "api: missing content-length"

	ErrUnknownError         = NewError(defaultErrorCode, defaultErrorMessage)
	ErrMethodNotAllowed     = NewError(methodNotAllowedErrorCode, methodNotAllowedErrorMessage)
	ErrNotFound             = NewError(notFoundErrorCode, notFoundErrorMessage)
	ErrLogExist             = NewError(logExistErrorCode, logExistErrorMessage)
	ErrLogNotFound          = NewError(logNotFoundErrorCode, logNotFoundErrorMessage)
	ErrLogNotAvailable      = NewError(logNotFoundErrorCode, logNotFoundErrorMessage)
	ErrMissingContentLength = NewError(missingLengthErrorCode, missingLengthErrorMessage)
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewError(code string, message string) (e *Error) {

	e = &Error{
		Code:    code,
		Message: message,
	}

	return e
}

func (e *Error) Error() (m string) {

	return e.Message
}

type ParamsError Error

func NewParamsError(err error) (e *ParamsError) {

	e = &ParamsError{
		Code:    paramsErrorCode,
		Message: fmt.Sprintf("api: params: %s", err.Error()),
	}

	return e
}
