package service

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"log"
)

type TransactionService struct {
	database.ITransactionDB
}

func NewTransactionService(db database.ITransactionDB) *TransactionService {
	service := &TransactionService{ITransactionDB: db}
	return service
}

// TODO: validate the coin whether spent or not
func (service *TransactionService) Validate(tx *model.Transaction) (uint64, error) {
	hash, err := validateHash[*model.Transaction](tx.Hash, tx)
	if err != nil {
		return 0, err
	}

	err = validateTimestamp(tx.Timestamp)
	if err != nil {
		return 0, err
	}

	existTx, err := service.GetTx(hash)
	if err != nil {
		return 0, err
	}
	if existTx != nil {
		return 0, errors.ErrTxExist
	}

	var totalInput uint64
	totalInput, err = service.validateInputs(tx)
	if err != nil {
		return 0, err
	}

	var totalOutput uint64
	totalOutput, err = service.validateOutputs(tx)
	if err != nil {
		return 0, err
	}

	if totalInput < totalOutput {
		return 0, errors.ErrTxInsufficientCoins
	}

	return totalInput - totalOutput, nil
}

func (service *TransactionService) SaveTxs(txs []*model.Transaction) {
	for _, tx := range txs {
		err := service.ITransactionDB.SaveTx(tx)
		if err != nil {
			log.Printf("save tx %x error", tx.Hash)
		}
	}
}

func (service *TransactionService) validateInputs(tx *model.Transaction) (uint64, error) {
	//TODO: empty inputs should fail
	if len(tx.Ins) != int(tx.InLen) {
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
	if prevTx == nil || prevTx.BlockHash == nil || len(prevTx.BlockHash) == 0 {
		return 0, errors.ErrPrevTxNotFound
	}
	if input.Index >= uint32(len(prevTx.Outs)) {
		return 0, errors.ErrInLenOutOfIndex
	}
	if prevTx.Timestamp.Compare(tx.Timestamp) >= 0 {
		return 0, errors.ErrTxTooLate
	}
	prevOutput := prevTx.Outs[input.Index]
	valid, err := cryptography.Verify(prevOutput.Pubkey, prevTx.Hash, input.Signature)
	if !valid || err != nil {
		return 0, errors.ErrTxSigInvalid
	}
	return prevOutput.Value, nil
}

func (service *TransactionService) validateOutputs(tx *model.Transaction) (uint64, error) {
	//TODO: empty outputs should fail
	if len(tx.Outs) != int(tx.OutLen) {
		return 0, errors.ErrOutLenMismatch
	}
	var total uint64 = 0
	for _, output := range tx.Outs {
		total += output.Value
	}
	return total, nil
}
