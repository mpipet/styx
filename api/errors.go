package api

var (
	defaultErrorCode          = "unknown_error"
	methodNotAllowedErrorCode = "method_not_allowed"
	notFoundErrorCode         = "not_found"

	defaultErrorMessage     = "api: unknown error"
	methodNotAllowedMessage = "api: method not allowed"
	notFoundMessage         = "api: not found"

	ErrUnknownError     = NewError(defaultErrorCode, defaultErrorMessage)
	ErrMethodNotAllowed = NewError(methodNotAllowedErrorCode, methodNotAllowedMessage)
	ErrNotFound         = NewError(notFoundErrorCode, notFoundMessage)
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
