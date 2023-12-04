package service

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"bytes"
)

type TransactionService struct {
	database.ITransactionDB
}

func NewTransactionService(db database.ITransactionDB) *TransactionService {
	service := &TransactionService{
		ITransactionDB: db,
	}
	return service
}

// TODO: test cases
func (service *TransactionService) ValidateOnChainTxs(txs []*model.Transaction, blockhash []byte, reward uint64) error {
	var totalFee uint64
	for i := 1; i < len(txs); i++ {
		fee, err := service.validate(txs[i], blockhash, false)
		if err != nil {
			return err
		}
		totalFee += fee
	}

	if err := service.validateCoinbase(txs[0], blockhash, totalFee+reward); err != nil {
		return err
	}
	return nil
}

func (service *TransactionService) ValidateOffChainTx(tx *model.Transaction) (uint64, error) {
	return service.validate(tx, nil, false)
}

func (service *TransactionService) validateCoinbase(tx *model.Transaction, blockhash []byte, val uint64) error {
	if _, err := service.validate(tx, blockhash, true); err != nil {
		return err
	}
	if tx.InLen != 0 {
		return errors.ErrTxCoinbaseInvalid
	}
	if tx.OutLen != 1 {
		return errors.ErrTxCoinbaseInvalid
	}
	if val != tx.Outs[0].Value {
		return errors.ErrTxCoinbaseInvalid
	}
	return nil
}

func (service *TransactionService) validate(tx *model.Transaction, blockhash []byte, coinbase bool) (uint64, error) {
	hash, err := validateHash[*model.Transaction](tx.Hash, tx)
	if err != nil {
		return 0, err
	}

	err = validateTimestamp(tx.Timestamp)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(tx.BlockHash, blockhash) {
		return 0, errors.ErrTxBlockHashInvalid
	}

	existTx, err := service.GetTx(hash)
	if err != nil {
		return 0, err
	}
	if existTx != nil {
		return 0, errors.ErrTxExist
	}

	var totalInput uint64
	totalInput, err = service.validateInputs(tx, coinbase)
	if err != nil {
		return 0, err
	}

	var totalOutput uint64
	totalOutput, err = service.validateOutputs(tx)
	if err != nil {
		return 0, err
	}

	if totalInput < totalOutput {
		return 0, errors.ErrTxNotEnoughValues
	}

	return totalInput - totalOutput, nil
}

func (service *TransactionService) validateInputs(tx *model.Transaction, coinbase bool) (uint64, error) {
	if len(tx.Ins) != int(tx.InLen) {
		return 0, errors.ErrInLenMismatch
	}
	if !coinbase && tx.InLen == 0 {
		return 0, errors.ErrInLenMismatch
	}

	var total uint64 = 0
	for _, input := range tx.Ins {
		val, err := service.validateInput(input, tx)
		if err != nil {
			return 0, err
		}
		total += val
	}
	return total, nil
}

func (service *TransactionService) validateInput(input *model.In, tx *model.Transaction) (uint64, error) {
	prevTx, err := service.GetTx(input.PrevHash)
	if err != nil {
		return 0, err
	}
	if prevTx == nil {
		return 0, errors.ErrPrevTxNotFound
	}
	if input.Index >= uint32(len(prevTx.Outs)) {
		return 0, errors.ErrInLenOutOfIndex
	}
	if prevTx.Timestamp.Compare(tx.Timestamp) >= 0 {
		return 0, errors.ErrInTooLate
	}

	input.PrevOut = prevTx.Outs[input.Index]

	valid, err := cryptography.Verify(input.PrevOut.Pubkey, prevTx.Hash, input.Signature)
	if !valid || err != nil {
		return 0, errors.ErrInSigInvalid
	}
	return input.PrevOut.Value, nil
}

func (service *TransactionService) validateOutputs(tx *model.Transaction) (uint64, error) {
	if len(tx.Outs) != int(tx.OutLen) {
		return 0, errors.ErrOutLenMismatch
	}
	var total uint64 = 0
	for _, output := range tx.Outs {
		total += output.Value
	}
	return total, nil
}
