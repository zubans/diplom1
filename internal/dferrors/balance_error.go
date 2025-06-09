package dferrors

import "errors"

var (
	ErrDuplicateWithdrawal = errors.New("duplicate withdrawal order")
	ErrInsufficientFunds   = errors.New("insufficient funds")
)
