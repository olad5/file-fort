package errors

import "errors"

const (
	ErrSomethingWentWrong = "something went wrong"
	ErrUnauthorized       = "unauthorized"
	ErrUserNotAdmin       = "unauthorized to perform this action"
	ErrInvalidJson        = "Invalid JSON"
	ErrMissingBody        = "missing body request"
)

var ErrInvalidID = errors.New("ID is not in its proper form")
