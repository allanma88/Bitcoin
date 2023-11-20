package service

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"log"
)

type TransactionService struct {
	utxo map[string]uint64
	database.ITransactionDB
}

func NewTransactionService(db database.ITransactionDB) *TransactionService {
	service := &TransactionService{
		ITransactionDB: db,
		utxo:           make(map[string]uint64),
	}
	return service
}

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
		return 0, errors.ErrTxNotEnoughValues
	}

	return totalInput - totalOutput, nil
}

func (service *TransactionService) ChainOnTxs(txs ...*model.Transaction) error {
	for _, tx := range txs {
		if tx.BlockHash == nil || len(tx.BlockHash) == 0 {
			return errors.ErrTxNotOnChain
		}

		for _, in := range tx.Ins {
			prevTx, err := service.GetOnChainTx(in.PrevHash)
			if err != nil {
				return err
			}
			if prevTx == nil {
				return errors.ErrPrevTxNotFound
			}

			out := prevTx.Outs[in.Index]
			if service.utxo[string(out.Pubkey)] < out.Value {
				return errors.ErrAccountNotEnoughValues
			}

			service.utxo[string(out.Pubkey)] -= out.Value
			if service.utxo[string(out.Pubkey)] == 0 {
				log.Printf("remove %x from uxto", out.Pubkey[:10])
				delete(service.utxo, string(out.Pubkey))
			}
		}

		for _, out := range tx.Outs {
			log.Printf("add %x to uxto", out.Pubkey[:10])
			service.utxo[string(out.Pubkey)] += out.Value
		}

		err := service.ITransactionDB.SaveOnChainTx(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *TransactionService) GetBalance(pubkey []byte) (uint64, bool) {
	val, ok := service.utxo[string(pubkey)]
	return val, ok
}

func (service *TransactionService) validateInputs(tx *model.Transaction) (uint64, error) {
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
	prevTx, err := service.GetOnChainTx(input.PrevHash)
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

	prevOutput := prevTx.Outs[input.Index]
	if service.utxo[string(prevOutput.Pubkey)] < prevOutput.Value {
		return 0, errors.ErrAccountNotEnoughValues
	}

	valid, err := cryptography.Verify(prevOutput.Pubkey, prevTx.Hash, input.Signature)
	if !valid || err != nil {
		return 0, errors.ErrInSigInvalid
	}
	return prevOutput.Value, nil
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
