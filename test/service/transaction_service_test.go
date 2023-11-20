package service

import (
	bcerrors "Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/service"
	"errors"
	"testing"
	"time"
)

func Test_ChainOnTxs_Not_On_Chain(t *testing.T) {
	tx := &model.Transaction{
		Ins:       []*model.In{},
		Outs:      []*model.Out{},
		BlockHash: []byte{},
	}
	formalizeTx(tx)

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	err := service.ChainOnTxs(tx)
	if !errors.Is(bcerrors.ErrTxNotOnChain, err) {
		t.Fatalf("on chain tx failed, expect: %v, actual %v", bcerrors.ErrTxNotOnChain, err)
	}
}

func Test_ChainOnTxs_No_PrevTx(t *testing.T) {
	_, tx := newTransactionPair(10, 8, time.Minute, nil, nil)

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	err := service.ChainOnTxs(tx)
	if !errors.Is(bcerrors.ErrPrevTxNotFound, err) {
		t.Fatalf("on chain tx failed, expect: %v, actual %v", bcerrors.ErrPrevTxNotFound, err)
	}
}

func Test_ChainOnTxs_PrevOut_Not_In_UXTO(t *testing.T) {
	prevTx, tx := newTransactionPair(10, 8, time.Minute, nil, nil)

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)
	err := service.SaveOnChainTx(prevTx)
	if err != nil {
		t.Fatalf("save prev tx error: %v", err)
	}

	err = service.ChainOnTxs(tx)
	if !errors.Is(bcerrors.ErrAccountNotEnoughValues, err) {
		t.Fatalf("on chain tx failed, expect: %v, actual %v", bcerrors.ErrAccountNotEnoughValues, err)
	}
}

func Test_ChainOnTxs_OK(t *testing.T) {
	var txVal uint64 = 8
	prevTx, tx := newTransactionPair(txVal+8, txVal, time.Minute, nil, nil)

	txdb := newTransactionDB()
	service := service.NewTransactionService(txdb)

	err := service.ChainOnTxs(prevTx, tx)
	if err != nil {
		t.Fatalf("put prevtx and tx on chain error: %v", err)
	}

	_, ok := service.GetBalance(prevTx.Outs[0].Pubkey)
	if ok {
		t.Fatalf("account %x of prev tx didn't remove from utxo", prevTx.Outs[0].Pubkey[:10])
	}

	val, _ := service.GetBalance(tx.Outs[0].Pubkey)
	if val != txVal {
		t.Fatalf("balance of account %x should be %d, acutally is %d", tx.Outs[0].Pubkey, txVal, val)
	}
}
