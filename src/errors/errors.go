package errors

import "errors"

var (
	ErrIdentityInvalid        = errors.New("invalid identity")
	ErrIdentityHashInvalid    = errors.New("invalid identity hash")
	ErrIdentityTooEarly       = errors.New("identity is too early")
	ErrTxExist                = errors.New("transaction already exists")
	ErrTxNotFound             = errors.New("transaction not found")
	ErrTxNotOnChain           = errors.New("transaction not on chain")
	ErrPrevTxNotFound         = errors.New("prev transaction not found")
	ErrTxNotEnoughValues      = errors.New("transaction not enough values")
	ErrInLenMismatch          = errors.New("transaction input length mismatch")
	ErrInLenOutOfIndex        = errors.New("transaction input out of index of prev transaction outputs")
	ErrInTooLate              = errors.New("transaction input is later than prev transaction")
	ErrInSigInvalid           = errors.New("transaction input signature invalid")
	ErrOutLenMismatch         = errors.New("transaction output length mismatch")
	ErrMerkleInvalid          = errors.New("invalid merkle tree")
	ErrBlockExist             = errors.New("block already exists")
	ErrBlockNotFound          = errors.New("block not found")
	ErrBlockNonceInvalid      = errors.New("invalid block nonce")
	ErrBlockContentInvalid    = errors.New("invalid block content")
	ErrBlockNoValidHash       = errors.New("no valid block hash")
	ErrServerCancelMining     = errors.New("the block is already mined")
	ErrAccountNotEnoughValues = errors.New("account not enough values")
)
