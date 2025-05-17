package apierr

import (
	"errors"
	"net/http"
)

var (
	NotFound         = errors.New("not found")
	InsufficientFund = errors.New("infufficient fund")
)

type JSON interface {
	error
	HTTPStatusCode() int
}

type json struct {
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *json) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *json) Unwrap() error {
	return e.Err
}

func (e *json) HTTPStatusCode() int {
	return e.StatusCode
}

func NewJSON(statusCode int, code, message string, originalErr error) JSON {
	return &json{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        originalErr,
	}
}

const (
	CodeValidationFailed    = "VALIDATION_FAILED"
	CodeNotFound            = "NOT_FOUND"
	CodeUnauthenticated     = "UNAUTHENTICATED"
	CodeForbidden           = "FORBIDDEN"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeConflict            = "CONFLICT"
	CodeUnprocessabled      = "UNPROCESSIABLED"
)

func BadRequest(message string) JSON {
	return NewJSON(http.StatusBadRequest, CodeValidationFailed, message, errors.New(message))
}

func InternalServer(message string) JSON {
	return NewJSON(http.StatusInternalServerError, CodeInternalServerError, message, errors.New(message))
}

func Unauthenticated() JSON {
	msg := "unauthenticated"
	return NewJSON(http.StatusUnauthorized, CodeUnauthenticated, msg, errors.New(msg))
}

func Unprocessable(message string) JSON {
	return NewJSON(http.StatusUnprocessableEntity, CodeUnprocessabled, message, errors.New(message))
}
