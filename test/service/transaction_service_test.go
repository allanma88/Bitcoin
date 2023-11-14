package service

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/database"
	bcerrors "Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/service"
	"Bitcoin/test"
	"bytes"
	"errors"

	"testing"
	"time"
)

func Test_Validate_Hash_Mismatch(t *testing.T) {
	tx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      []*model.Out{},
		Timestamp: time.Now(),
	}
	err := formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	tx.Hash, err = cryptography.Hash("whatever")
	if err != nil {
		t.Fatalf("hash transaction  error: %s", err)
	}

	service := &service.TransactionService{}
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrIdentityHashInvalid) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrIdentityHashInvalid, err)
	}
}

func Test_Validate_Tx_Exists(t *testing.T) {
	tx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      []*model.Out{},
		Timestamp: time.Now(),
	}
	err := formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(tx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxExist) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxExist, err)
	}
}

func Test_Validate_Time_Too_Early(t *testing.T) {
	future := time.Now().Add(2*time.Hour + time.Minute)
	tx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      []*model.Out{},
		Timestamp: future,
	}
	err := formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrIdentityTooEarly) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrIdentityTooEarly, err)
	}
}

func Test_Validate_Ins_Len_Mismatch(t *testing.T) {
	tx := &model.Transaction{
		InLen: 3,
		Ins:   []*model.In{{}, {}},
		Outs:  []*model.Out{},
	}
	err := formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrInLenMismatch) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrInLenMismatch, err)
	}
}

func Test_Validate_Input_PrevTx_Not_Found(t *testing.T) {
	ins, err := newIns(nil, nil, 0)
	if err != nil {
		t.Fatalf("new ins error: %s", err)
	}

	tx := &model.Transaction{
		Ins:       ins,
		Outs:      []*model.Out{},
		Timestamp: time.Now(),
	}
	err = formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrPrevTxNotFound) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrTxNotFound, err)
	}
}

func Test_Validate_Input_PrevTx_Out_Chain_Found(t *testing.T) {
	prevTx, tx, err := newTransactionPair(10, 0, time.Minute, []byte{})
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrPrevTxNotFound) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrTxNotFound, err)
	}
}

func Test_Validate_Input_Time_Same_As_PrevTx(t *testing.T) {
	prevTx, tx, err := newTransactionPair(10, 0, 0, nil)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxTooLate, err)
	}
}

func Test_Validate_Input_Time_Too_Late(t *testing.T) {
	prevTx, tx, err := newTransactionPair(10, 0, -1*time.Minute, nil)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxTooLate, err)
	}
}

func Test_Validate_In_Sig_Mismatch(t *testing.T) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys error: %s", err)
	}

	prevOuts := newOuts(pubkey, 10)
	prevTx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      prevOuts,
		Timestamp: time.Now(),
	}
	err = formalizeTx(prevTx)
	if err != nil {
		t.Fatalf("formalize prev transaction error: %s", err)
	}

	ins, err := newIns(privkey, prevTx.Hash, 0)
	if err != nil {
		t.Fatalf("new ins error: %s", err)
	}
	ins[0].Signature = []byte{}

	tx := &model.Transaction{
		Ins:       ins,
		Outs:      []*model.Out{},
		Timestamp: time.Now().Add(time.Minute),
	}
	err = formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxSigInvalid) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxSigInvalid, err)
	}
}

func Test_Validate_Outs_Len_Not_Match(t *testing.T) {
	tx := &model.Transaction{
		OutLen: 3,
		Ins:    []*model.In{},
		Outs:   []*model.Out{{}, {}},
	}
	err := formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrOutLenMismatch) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrOutLenMismatch, err)
	}
}

func Test_Validate_Output_Value_Too_Large(t *testing.T) {
	var prevVal uint64 = 10
	prevTx, tx, err := newTransactionPair(prevVal, prevVal+1, time.Minute, nil)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxInsufficientCoins) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxInsufficientCoins, err)
	}
}

func Test_Validate_Success(t *testing.T) {
	var totalInput uint64 = 10
	var totalOutput uint64 = 6
	prevTx, tx, err := newTransactionPair(totalInput, totalOutput, time.Minute, nil)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB()
	txdb.SaveOnChainTx(prevTx)
	service := service.NewTransactionService(txdb)
	fee, err := service.Validate(tx)
	if err != nil {
		t.Fatalf("transaction validate error: %s", err)
	}

	if fee != totalInput-totalOutput {
		t.Fatalf("transaction fee, expect: %d, actual: %d", totalInput-totalOutput, fee)
	}
}

func Test_Validate_Hash_Not_Change(t *testing.T) {
	tx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      []*model.Out{},
		Timestamp: time.Now(),
	}
	err := formalizeTx(tx)
	if err != nil {
		t.Fatalf("formalize transaction error: %s", err)
	}

	originalHash := tx.Hash

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if err != nil {
		t.Fatalf("validate transaction error: %s", err)
	}

	if bytes.Equal(originalHash, tx.Hash) {
		t.Log("transaction hash didn't changed after serialize")
	} else {
		t.Fatalf("transaction hash is changed from [%x] to [%x]", originalHash, tx.Hash)
	}
}

func newIns(privkey, prevHash []byte, index uint32) ([]*model.In, error) {
	var err error
	if privkey == nil {
		privkey, _, err = test.NewKeys()
		if err != nil {
			return nil, err
		}
	}

	if prevHash == nil {
		prevHash, err = cryptography.Hash("whatever")
		if err != nil {
			return nil, err
		}
	}

	sig, err := cryptography.Sign(privkey, prevHash)
	if err != nil {
		return nil, err
	}

	ins := []*model.In{
		{
			PrevHash:  prevHash,
			Index:     index,
			Signature: sig,
		},
	}
	return ins, nil
}

func newTransactionPair(prevVal, val uint64, duration time.Duration, prevBlockHash []byte) (*model.Transaction, *model.Transaction, error) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	prevOuts := newOuts(pubkey, prevVal)
	prevTx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      prevOuts,
		Timestamp: now,
		BlockHash: prevBlockHash,
	}
	err = formalizeTx(prevTx)
	if err != nil {
		return nil, nil, err
	}

	ins, err := newIns(privkey, prevTx.Hash, 0)
	if err != nil {
		return nil, nil, err
	}
	outs := newOuts(pubkey, val)
	tx := &model.Transaction{
		Ins:       ins,
		Outs:      outs,
		Timestamp: now.Add(duration),
		BlockHash: []byte{},
	}
	err = formalizeTx(tx)
	if err != nil {
		return nil, nil, err
	}

	return prevTx, tx, nil
}

func newOuts(pubkey []byte, val uint64) []*model.Out {
	outs := make([]*model.Out, 0)
	if val > 0 {
		out := &model.Out{
			Pubkey: pubkey,
			Value:  val,
		}
		outs = append(outs, out)
	}
	return outs
}

func formalizeTx(tx *model.Transaction) error {
	if tx.InLen == 0 {
		tx.InLen = uint32(len(tx.Ins))
	}
	if tx.OutLen == 0 {
		tx.OutLen = uint32(len(tx.Outs))
	}
	if tx.BlockHash == nil {
		blockHash, err := cryptography.Hash("block")
		if err != nil {
			return err
		}
		tx.BlockHash = blockHash
	}

	if tx.Timestamp.IsZero() {
		tx.Timestamp = time.Now()
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		return err
	}

	tx.Hash = hash
	return nil
}

func newTransactionDB(txs ...*model.Transaction) database.ITransactionDB {
	basedb := newTestBaseDB[model.Transaction]()
	txdb := &database.TransactionDB{IBaseDB: basedb}
	return txdb
}
