package model

import "errors"

var (
	ErrTxExist             = errors.New("transaction already exists")
	ErrTxHashInvalid       = errors.New("transaction hash invalid")
	ErrTxTooLate           = errors.New("transaction is later than prev transaction")
	ErrTxTooEarly          = errors.New("transaction is too early")
	ErrTxSigInvalid        = errors.New("transaction signature invalid")
	ErrTxNotFound          = errors.New("transaction not found")
	ErrTxInsufficientCoins = errors.New("transaction insufficient coins")
	ErrInLenMismatch       = errors.New("transaction input length mismatch")
	ErrInLenOutOfIndex     = errors.New("transaction input out of index of prev transaction outputs")
	ErrOutLenMismatch      = errors.New("transaction output length mismatch")
)
