package service

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"bytes"
)

type TransactionService struct {
	database.IBlockDB
	utxo map[string]uint64
}

type GetTxFunc func([]byte) *model.Transaction

func NewTransactionService(db database.IBlockDB, utxo map[string]uint64) *TransactionService {
	service := &TransactionService{
		IBlockDB: db,
		utxo:     utxo,
	}
	return service
}

// TODO: test cases
func (s *TransactionService) ValidateOnChainTxs(txs []*model.Transaction, blockhash []byte, reward uint64, isMainChain bool) error {
	txmap := make(map[string]*model.Transaction)
	f := func(hash []byte) *model.Transaction {
		return txmap[string(hash)]
	}

	var utxo map[string]uint64 = nil
	if isMainChain {
		utxo = s.utxo
	}

	var totalFee uint64 = 0
	for _, tx := range txs {
		if err := s.validateTx(tx, blockhash, false, utxo, f); err != nil {
			return err
		}
		txmap[string(tx.Hash)] = tx
		totalFee += tx.Fee
	}

	if err := s.validateCoinbase(txs[0], blockhash, totalFee+reward); err != nil {
		return err
	}
	return nil
}

func (s *TransactionService) ValidateTx(tx *model.Transaction, f GetTxFunc) error {
	return s.validateTx(tx, nil, false, s.utxo, f)
}

func (s *TransactionService) validateCoinbase(tx *model.Transaction, blockhash []byte, val uint64) error {
	if err := s.validateTx(tx, blockhash, true, nil, nil); err != nil {
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

func (s *TransactionService) validateTx(tx *model.Transaction, blockhash []byte, coinbase bool, utxo map[string]uint64, f GetTxFunc) error {
	hash, err := validateHash[*model.Transaction](tx.Hash, tx)
	if err != nil {
		return err
	}

	err = validateTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	if !bytes.Equal(tx.BlockHash, blockhash) {
		return errors.ErrTxBlockHashInvalid
	}

	existTx, err := s.GetTx(hash)
	if err != nil {
		return err
	}
	if existTx != nil {
		return errors.ErrTxExist
	}

	var totalInput uint64
	totalInput, err = s.validateInputs(tx, coinbase, utxo, f)
	if err != nil {
		return err
	}

	var totalOutput uint64
	totalOutput, err = s.validateOutputs(tx)
	if err != nil {
		return err
	}

	if totalInput < totalOutput {
		return errors.ErrTxNotEnoughValues
	}
	tx.Fee = totalInput - totalOutput
	return nil
}

func (s *TransactionService) validateInputs(tx *model.Transaction, coinbase bool, utxo map[string]uint64, f GetTxFunc) (uint64, error) {
	if len(tx.Ins) != int(tx.InLen) {
		return 0, errors.ErrInLenMismatch
	}
	if !coinbase && tx.InLen == 0 {
		return 0, errors.ErrInLenMismatch
	}

	var total uint64 = 0
	for _, input := range tx.Ins {
		val, err := s.validateInput(input, tx, utxo, f)
		if err != nil {
			return 0, err
		}
		total += val
	}
	return total, nil
}

func (s *TransactionService) validateInput(input *model.In, tx *model.Transaction, utxo map[string]uint64, f GetTxFunc) (uint64, error) {
	prevTx, err := s.GetTx(input.PrevHash)
	if err != nil {
		return 0, err
	}
	if prevTx == nil {
		prevTx = f(input.PrevHash)
		if prevTx == nil {
			return 0, errors.ErrPrevTxNotFound
		}
	}
	if input.Index >= uint32(len(prevTx.Outs)) {
		return 0, errors.ErrInLenOutOfIndex
	}
	if prevTx.Timestamp.Compare(tx.Timestamp) > 0 {
		return 0, errors.ErrInTooLate
	}

	input.PrevOut = prevTx.Outs[input.Index].DeepClone()

	if utxo != nil && utxo[string(input.PrevOut.Pubkey)] < input.PrevOut.Value {
		return 0, errors.ErrAccountNotEnoughValues
	}

	valid, err := cryptography.Verify(input.PrevOut.Pubkey, prevTx.Hash, input.Signature)
	if !valid || err != nil {
		return 0, errors.ErrInSigInvalid
	}
	return input.PrevOut.Value, nil
}

func (s *TransactionService) validateOutputs(tx *model.Transaction) (uint64, error) {
	if len(tx.Outs) != int(tx.OutLen) {
		return 0, errors.ErrOutLenMismatch
	}
	var total uint64 = 0
	for _, output := range tx.Outs {
		total += output.Value
	}
	return total, nil
}
