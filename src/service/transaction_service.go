package service

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/db"
	"Bitcoin/src/model"
	"bytes"
	"time"
)

type TransactionService struct {
	db.ITransactionDB
}

func NewTransactionService(db db.ITransactionDB) *TransactionService {
	service := &TransactionService{ITransactionDB: db}
	return service
}

func (service *TransactionService) Validate(tx *model.Transaction) error {
	hash, err := validateHash(tx)
	if err != nil {
		return err
	}

	existTx, err := service.GetTx(hash)
	if existTx != nil && err == nil {
		return model.ErrTxExist
	}

	if tx.Timestamp.AsTime().Compare(time.Now().Add(2*time.Hour)) >= 0 {
		return model.ErrTxTooEarly
	}

	var totalInput uint64
	totalInput, err = service.validateInputs(tx)
	if err != nil {
		return err
	}

	var totalOutput uint64
	totalOutput, err = service.validateOutputs(tx)
	if err != nil {
		return err
	}

	if totalInput < totalOutput {
		return model.ErrTxInsufficientCoins
	}

	return nil
}

func validateHash(tx *model.Transaction) ([]byte, error) {
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, model.ErrTxHashInvalid
	}
	if !bytes.Equal(hash, tx.Id) {
		return nil, model.ErrTxHashInvalid
	}
	return hash, nil
}

func (service *TransactionService) validateInputs(tx *model.Transaction) (uint64, error) {
	//TODO: empty inputs should fail
	if len(tx.Ins) != int(tx.InLen) {
		return 0, model.ErrInLenMismatch
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
		return 0, model.ErrTxNotFound
	}
	if input.Index >= uint32(len(prevTx.Outs)) {
		return 0, model.ErrInLenOutOfIndex
	}
	if prevTx.Timestamp.AsTime().Compare(tx.Timestamp.AsTime()) >= 0 {
		return 0, model.ErrTxTooLate
	}
	prevOutput := prevTx.Outs[input.Index]
	valid, err := cryptography.Verify(prevOutput.Pubkey, prevTx.Id, input.Signature)
	if !valid || err != nil {
		return 0, model.ErrTxSigInvalid
	}
	return prevOutput.Value, nil
}

func (service *TransactionService) validateOutputs(tx *model.Transaction) (uint64, error) {
	//TODO: empty outputs should fail
	if len(tx.Outs) != int(tx.OutLen) {
		return 0, model.ErrOutLenMismatch
	}
	var total uint64 = 0
	for _, output := range tx.Outs {
		total += output.Value
	}
	return total, nil
}
