package errs

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Error struct {
	kind error
	msg  string
}

func (e *Error) Error() string { return e.msg }

func (e *Error) Unwrap() error { return e.kind }

func New(kind error, msg string) error {
	return &Error{kind: kind, msg: msg}
}

func BadRequest(msg string) error {
	return New(ErrInvalidInput, msg)
}

func NotFound(msg string) error {
	return New(ErrNotFound, msg)
}

func Unauthorized(msg string) error {
	return New(ErrUnauthorized, msg)
}

func Forbidden(msg string) error {
	return New(ErrForbidden, msg)
}

func Conflict(msg string) error {
	return New(ErrConflict, msg)
}

func InvalidCredentials(msg string) error {
	return New(ErrInvalidCredentials, msg)
}
