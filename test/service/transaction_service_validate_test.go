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
	"log"

	"testing"
	"time"
)

func Test_Validate_Hash_Mismatch(t *testing.T) {
	tx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      []*model.Out{},
		Timestamp: time.Now(),
	}
	formalizeTx(tx)

	hash, err := cryptography.Hash("whatever")
	if err != nil {
		t.Fatalf("hash transaction error: %s", err)
	}
	tx.Hash = hash

	service := &service.TransactionService{}
	_, err = service.ValidateOffChainTx(tx)
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
	formalizeTx(tx)

	txdb := newTransactionDB()
	service := newTransactionService(txdb)
	service.SaveOffChainTx(tx)

	_, err := service.ValidateOffChainTx(tx)
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
	formalizeTx(tx)

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err := service.ValidateOffChainTx(tx)
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
	formalizeTx(tx)

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err := service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrInLenMismatch) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrInLenMismatch, err)
	}
}

func Test_Validate_Input_PrevTx_Not_Found(t *testing.T) {
	_, tx := newTransactionPair(10, 8, time.Minute, nil, []byte{})

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	_, err := service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrPrevTxNotFound) {
		t.Fatalf("transaction validate failed, expect: %s, actual: %s", bcerrors.ErrTxNotFound, err)
	}
}

func Test_Validate_Input_Time_Same_As_PrevTx(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 0, 0, nil, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err := service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrInTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrInTooLate, err)
	}
}

func Test_Validate_Input_Time_Too_Late(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 0, -1*time.Minute, nil, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err := service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrInTooLate) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrInTooLate, err)
	}
}

func Test_Validate_Input_PrevTx_Out_Chain_Not_Found(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 0, time.Minute, []byte{}, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb)

	err := service.SaveOnChainTx(prevTx)
	if err != nil {
		t.Fatalf("save prev tx on chain error: %v", err)
	}

	_, err = service.ValidateOffChainTx(tx)
	if err != nil {
		t.Fatalf("transaction validate failed: %v", err)
	}
}

func Test_Validate_In_Sig_Mismatch(t *testing.T) {
	privkey, pubkey := test.NewKeys()

	blockHash, err := cryptography.Hash("block")
	if err != nil {
		log.Fatalf("compute hash error: %v", err)
	}

	prevOuts := newOuts(pubkey, 10)
	prevTx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      prevOuts,
		Timestamp: time.Now(),
		BlockHash: blockHash,
	}
	formalizeTx(prevTx)

	in := newIn(privkey, prevTx, 0)
	in.Signature = []byte{}

	tx := &model.Transaction{
		Ins:       []*model.In{in},
		Outs:      []*model.Out{},
		Timestamp: time.Now().Add(time.Minute),
	}
	formalizeTx(tx)

	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err = service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrInSigInvalid) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrInSigInvalid, err)
	}
}

func Test_Validate_Outs_Len_Not_Match(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 8, time.Minute, nil, []byte{})

	tx.OutLen = uint32(len(tx.Outs)) + 1
	formalizeTx(tx)

	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)
	_, err := service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrOutLenMismatch) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrOutLenMismatch, err)
	}
}

func Test_Validate_Outs_Zero_Len(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 0, time.Minute, nil, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err := service.ValidateOffChainTx(tx)
	if err != nil {
		t.Fatalf("transaction validate failed, expect success, actual %v", err)
	}
}

func Test_Validate_Output_Value_Too_Small(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 0, time.Minute, nil, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err := service.ValidateOffChainTx(tx)
	if err != nil {
		t.Fatalf("transaction validate failed, expect: success, actual %s", err)
	}
}

func Test_Validate_Output_Value_Too_Large(t *testing.T) {
	var prevVal uint64 = 10
	prevTx, tx := newTransactionPair(prevVal, prevVal+1, time.Minute, nil, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err := service.ValidateOffChainTx(tx)
	if !errors.Is(err, bcerrors.ErrTxNotEnoughValues) {
		t.Fatalf("transaction validate failed, expect: %s, actual %s", bcerrors.ErrTxNotEnoughValues, err)
	}
}

func Test_Validate_Success(t *testing.T) {
	var totalInput uint64 = 10
	var totalOutput uint64 = 6

	prevTx, tx := newTransactionPair(totalInput, totalOutput, time.Minute, nil, []byte{})
	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	fee, err := service.ValidateOffChainTx(tx)
	if err != nil {
		t.Fatalf("transaction validate error: %s", err)
	}

	if fee != totalInput-totalOutput {
		t.Fatalf("transaction fee, expect: %d, actual: %d", totalInput-totalOutput, fee)
	}
}

func Test_Validate_Hash_Not_Change(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 8, time.Minute, nil, []byte{})

	originalHash := tx.Hash

	txdb := newTransactionDB()
	service := newTransactionService(txdb, prevTx)

	_, err := service.ValidateOffChainTx(tx)
	if err != nil {
		t.Fatalf("validate transaction error: %s", err)
	}

	if bytes.Equal(originalHash, tx.Hash) {
		t.Log("transaction hash didn't changed after serialize")
	} else {
		t.Fatalf("transaction hash is changed from [%x] to [%x]", originalHash, tx.Hash)
	}
}

func newTransactionPair(prevVal, val uint64, duration time.Duration, prevBlockHash, blockHash []byte) (*model.Transaction, *model.Transaction) {
	blockhash, err := cryptography.Hash("block")
	if err != nil {
		log.Fatalf("compute hash error: %v", err)
	}
	if prevBlockHash == nil {
		prevBlockHash = blockhash
	}
	if blockHash == nil {
		blockHash = blockhash
	}

	prevPrivkey, prevPubkey := test.NewKeys()

	now := time.Now()
	prevOuts := newOuts(prevPubkey, prevVal)
	prevTx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      prevOuts,
		Timestamp: now,
		BlockHash: prevBlockHash,
	}
	formalizeTx(prevTx)

	_, pubkey := test.NewKeys()
	in := newIn(prevPrivkey, prevTx, 0)

	outs := newOuts(pubkey, val)
	tx := &model.Transaction{
		Ins:       []*model.In{in},
		Outs:      outs,
		Timestamp: now.Add(duration),
		BlockHash: blockHash,
	}
	formalizeTx(tx)

	return prevTx, tx
}

func newIn(privkey []byte, prevTx *model.Transaction, index uint32) *model.In {
	sig, err := cryptography.Sign(privkey, prevTx.Hash)
	if err != nil {
		log.Fatalf("sign prev hash error: %v", err)
	}

	in := &model.In{
		PrevHash:  prevTx.Hash,
		Index:     index,
		Signature: sig,
		PrevOut:   prevTx.Outs[index],
	}
	return in
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

func formalizeTx(tx *model.Transaction) {
	if tx.InLen == 0 {
		tx.InLen = uint32(len(tx.Ins))
	}
	if tx.OutLen == 0 {
		tx.OutLen = uint32(len(tx.Outs))
	}

	if tx.Timestamp.IsZero() {
		tx.Timestamp = time.Now()
	}

	hash, err := tx.ComputeHash()
	if err != nil {
		log.Fatalf("compute tx hash error: %v", err)
	}

	tx.Hash = hash
}

func newTransactionDB(txs ...*model.Transaction) database.ITransactionDB {
	basedb := newTestBaseDB[model.Transaction]()
	txdb := &database.TransactionDB{IBaseDB: basedb}
	return txdb
}

func newTransactionService(txdb database.ITransactionDB, txs ...*model.Transaction) *service.TransactionService {
	service := service.NewTransactionService(txdb)
	for _, tx := range txs {
		err := service.ITransactionDB.SaveOnChainTx(tx)
		if err != nil {
			log.Fatalf("put tx %x on chain error: %v", tx.Hash, err)
		}
	}
	return service
}
