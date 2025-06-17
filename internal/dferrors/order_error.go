package dferrors

import "errors"

var (
	ErrInvalidNumber = errors.New("invalid order number")
	ErrOrderConflict = errors.New("order conflict")
)
