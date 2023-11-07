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
	tx, err := newTransaction4([]*model.In{}, []*model.Out{}, time.Now())
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
	tx, err := newTransaction4([]*model.In{}, []*model.Out{}, time.Now())
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
	tx, err := newTransaction4([]*model.In{}, []*model.Out{}, future)
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
	tx, err := newTransaction_InLen_Error(ins, []*model.Out{}, time.Now())
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
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys err: %s", err)
	}

	now := time.Now()
	prevTx, err := newTransaction1(pubkey, 10, now)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	prevHash, err := cryptography.Hash("whatever")
	if err != nil {
		t.Fatalf("create hash error: %s", err)
	}

	tx, err := newTransaction2(privkey, prevHash, 0, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxNotFound) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrTxNotFound, err)
	}
}

func Test_Validate_Input_Time_Same_As_PrevTx(t *testing.T) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys err: %s", err)
	}

	now := time.Now()
	prevTx, err := newTransaction1(pubkey, 10, now)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	tx, err := newTransaction2(privkey, prevTx.Hash, 0, now)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxTooLate, err)
	}
}

func Test_Validate_Input_Time_Too_Late(t *testing.T) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys err: %s", err)
	}

	now := time.Now()
	prevTx, err := newTransaction1(pubkey, 10, now)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	tx, err := newTransaction2(privkey, prevTx.Hash, 0, now.Add(-1*time.Minute))
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxTooLate, err)
	}
}

func Test_Validate_In_Sig_Mismatch(t *testing.T) {
	_, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys err: %s", err)
	}

	prevTx, err := newTransaction1(pubkey, 10, time.Now())
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	tx, err := newTransaction_Sig_Error(prevTx.Hash, 0, time.Now())
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxSigInvalid) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxSigInvalid, err)
	}
}

func Test_Validate_Outs_Len_Not_Match(t *testing.T) {
	var err error
	ins := []*model.In{}
	outs := []*model.Out{{}, {}}

	tx, err := newTransaction_OutLen_Error(ins, outs, time.Now())
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
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys error: %s", err)
	}

	var val uint64 = 10
	prevTx, err := newTransaction1(pubkey, val, time.Now())
	if err != nil {
		t.Fatalf("new prev transaction error: %s", err)
	}

	tx, err := newTransaction3(privkey, pubkey, prevTx.Hash, 0, val+1, time.Now())
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	txdb := newTransactionDB(prevTx)
	service := service.NewTransactionService(txdb)
	_, err = service.Validate(tx)
	if !errors.Is(err, bcerrors.ErrTxInsufficientCoins) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxInsufficientCoins, err)
	}
}

func Test_Validate_Success(t *testing.T) {
	privkey, pubkey, err := test.NewKeys()
	if err != nil {
		t.Fatalf("new keys err: %s", err)
	}

	now := time.Now()
	var totalInput uint64 = 10
	prevTx, err := newTransaction1(pubkey, totalInput, now)
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
	}

	var totalOutput uint64 = 6
	tx, err := newTransaction3(privkey, pubkey, prevTx.Hash, 0, totalOutput, time.Now())
	if err != nil {
		t.Fatalf("new transaction error: %s", err)
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
	tx, err := newTransaction4([]*model.In{}, []*model.Out{}, time.Now())
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

func newTransaction1(pubkey []byte, val uint64, timestamp time.Time) (*model.Transaction, error) {
	out := &model.Out{
		Pubkey: pubkey,
		Value:  val,
	}
	outs := []*model.Out{out}
	tx := &model.Transaction{
		InLen:     0,
		OutLen:    uint32(len(outs)),
		Ins:       []*model.In{},
		Outs:      outs,
		Timestamp: timestamp,
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction2(privkey, prevHash []byte, index uint32, timestamp time.Time) (*model.Transaction, error) {
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

	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    0,
		Ins:       ins,
		Outs:      []*model.Out{},
		Timestamp: timestamp,
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction3(privkey, pubkey, prevHash []byte, index uint32, val uint64, timestamp time.Time) (*model.Transaction, error) {
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

	outs := []*model.Out{{
		Pubkey: pubkey,
		Value:  val,
	}}

	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamp,
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction4(ins []*model.In, outs []*model.Out, timestamp time.Time) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamp,
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction_InLen_Error(ins []*model.In, outs []*model.Out, timestamp time.Time) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     uint32(len(ins)) - 1,
		OutLen:    uint32(len(outs)),
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamp,
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction_OutLen_Error(ins []*model.In, outs []*model.Out, timestamp time.Time) (*model.Transaction, error) {
	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    uint32(len(outs)) - 1,
		Ins:       ins,
		Outs:      outs,
		Timestamp: timestamp,
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash
	return tx, nil
}

func newTransaction_Sig_Error(prevHash []byte, index uint32, timestamp time.Time) (*model.Transaction, error) {
	ins := []*model.In{
		{
			PrevHash:  prevHash,
			Index:     index,
			Signature: []byte{},
		},
	}

	tx := &model.Transaction{
		InLen:     uint32(len(ins)),
		OutLen:    0,
		Ins:       ins,
		Outs:      []*model.Out{},
		Timestamp: timestamp,
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
