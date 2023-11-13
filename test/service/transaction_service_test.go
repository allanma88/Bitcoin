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
	tx, err := newTransaction1([]*model.In{}, []*model.Out{}, time.Now(), nil)
	if err != nil {
		t.Fatalf("new transaction 4 error: %s", err)
	}

	tx.Hash, err = cryptography.Hash("whatever")
	if err != nil {
		t.Fatalf("transaction hash error: %s", err)
	}

	service := &service.TransactionService{}
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrIdentityHashInvalid) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrIdentityHashInvalid, err)
	}
}

func Test_Validate_Tx_Exists(t *testing.T) {
	tx, err := newTransaction1([]*model.In{}, []*model.Out{}, time.Now(), nil)
	if err != nil {
		t.Fatalf("new transaction 4 error: %s", err)
	}

	txdb := newTransactionDB(tx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxExist) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxExist, err)
	}
}

func Test_Validate_Time_Too_Early(t *testing.T) {
	future := time.Now().Add(2*time.Hour + time.Minute)
	tx, err := newTransaction1([]*model.In{}, []*model.Out{}, future, nil)
	if err != nil {
		t.Fatalf("new transaction 4 error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrIdentityTooEarly) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrIdentityTooEarly, err)
	}
}

func Test_Validate_Ins_Len_Mismatch(t *testing.T) {
	ins := []*model.In{{}, {}}
	tx, err := newTransaction2(ins, []*model.Out{}, 3, 1)
	if err != nil {
		t.Fatalf("new transaction 4 error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrInLenMismatch) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrInLenMismatch, err)
	}
}

func Test_Validate_Input_PrevTx_Not_Found(t *testing.T) {
	privkey, _, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys err: %s", err)
	}

	prevHash, err := cryptography.Hash("whatever")
	if err != nil {
		t.Fatalf("create hash error: %s", err)
	}

	ins, err := newIns(privkey, prevHash, 0)
	if err != nil {
		t.Fatalf("new ins error: %s", err)
	}

	tx, err := newTransaction1(ins, []*model.Out{}, time.Now(), nil)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrPrevTxNotFound) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrTxNotFound, err)
	}
}

func Test_Validate_Input_PrevTx_Outof_Chain_Found(t *testing.T) {
	prevTx, tx, err := newTransactionPair(0, 10, 0)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrPrevTxNotFound) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrTxNotFound, err)
	}
}

func Test_Validate_Input_Time_Same_As_PrevTx(t *testing.T) {
	prevTx, tx, err := newTransactionPair(0, 10, 0)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	tx.Timestamp = prevTx.Timestamp

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxTooLate, err)
	}
}

func Test_Validate_Input_Time_Too_Late(t *testing.T) {
	prevTx, tx, err := newTransactionPair(0, 10, 0)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	tx.Timestamp = prevTx.Timestamp.Add(-1 * time.Minute)
	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxTooLate, err)
	}
}

func Test_Validate_In_Sig_Mismatch(t *testing.T) {
	prevTx, tx, err := newTransactionPair(0, 10, 0)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	tx.Ins[0].Signature = []byte{}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxSigInvalid) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxSigInvalid, err)
	}
}

func Test_Validate_Outs_Len_Not_Match(t *testing.T) {
	ins := []*model.In{}
	outs := []*model.Out{{}, {}}

	tx, err := newTransaction2(ins, outs, 0, 3)
	if err != nil {
		t.Fatalf("new transaction 4 error: %s", err)
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
	prevTx, tx, err := newTransactionPair(0, prevVal, prevVal+1)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxInsufficientCoins) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxInsufficientCoins, err)
	}
}

func Test_Validate_Success(t *testing.T) {
	var totalInput uint64 = 10
	var totalOutput uint64 = 6
	prevTx, tx, err := newTransactionPair(0, totalInput, totalOutput)
	if err != nil {
		t.Fatalf("new transaction pair error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
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
	tx, err := newTransaction1([]*model.In{}, []*model.Out{}, time.Now(), nil)
	if err != nil {
		t.Fatalf("new transaction 4 error: %s", err)
	}

	originalHash := tx.Hash

	txdb := newTransactionDB(tx)
	service := service.NewTransactionService(txdb)
	service.Validate(tx)
	if bytes.Equal(originalHash, tx.Hash) {
		t.Log("transaction hash didn't changed after serialize")
	} else {
		t.Fatalf("transaction hash is changed from [%x] to [%x]", originalHash, tx.Hash)
	}
}

func newIns(privkey, prevHash []byte, index uint32) ([]*model.In, error) {
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

func newTransactionPair(index uint32, prevVal, val uint64) (*model.Transaction, *model.Transaction, error) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		return nil, nil, err
	}

	blockHash, err := cryptography.Hash("block")
	if err != nil {
		return nil, nil, err
	}

	prevOuts := newOuts(pubkey, prevVal)
	prevTx, err := newTransaction1([]*model.In{}, prevOuts, time.Now(), blockHash)
	if err != nil {
		return nil, nil, err
	}

	ins, err := newIns(privkey, prevTx.Hash, 0)
	if err != nil {
		return nil, nil, err
	}
	outs := newOuts(pubkey, val)
	tx, err := newTransaction1(ins, outs, time.Now(), nil)
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

func newTransaction1(ins []*model.In, outs []*model.Out, timestamp time.Time, blockHash []byte) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamp,
		BlockHash: blockHash,
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction2(ins []*model.In, outs []*model.Out, inlen, outlen uint32) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     inlen,
		OutLen:    outlen,
		Ins:       ins,
		Outs:      outs,
		Timestamp: time.Now(),
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}

	tx.Hash = hash
	return tx, nil
}

func newTransactionDB(txs ...*model.Transaction) database.ITransactionDB {
	basedb := newTestBaseDB[model.Transaction]()
	txdb := &database.TransactionDB{IBaseDB: basedb}
	for _, tx := range txs {
		txdb.SaveTx(tx)
	}
	return txdb
}
