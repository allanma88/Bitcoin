package service

import (
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
)

//TODO: test cases

type UtxoService struct {
	utxo map[string]uint64
}

func NewUtxoService() *UtxoService {
	return &UtxoService{utxo: make(map[string]uint64)}
}

func (service *UtxoService) GetBalance(pubkey []byte) uint64 {
	val := service.utxo[string(pubkey)]
	return val
}

func (service *UtxoService) RollbackBalances(txs []*model.Transaction) error {
	utxo := make(map[string]int64)
	for _, tx := range txs {
		for _, in := range tx.Ins {
			utxo[string(in.PrevOut.Pubkey)] += int64(in.PrevOut.Value)
		}

		for _, out := range tx.Outs {
			utxo[string(out.Pubkey)] -= int64(out.Value)
		}
	}
	for addr, val := range utxo {
		if int64(service.utxo[addr])+val < 0 {
			return errors.ErrAccountNotEnoughValues
		}
	}
	for addr, val := range utxo {
		service.utxo[addr] += uint64(val)
		if service.utxo[addr] == 0 {
			delete(service.utxo, addr)
		}
	}
	return nil
}

// TODO: test case: apply all or none
func (service *UtxoService) ApplyBalances(txs []*model.Transaction) error {
	utxo := make(map[string]int64)
	for _, tx := range txs {
		for _, in := range tx.Ins {
			utxo[string(in.PrevOut.Pubkey)] -= int64(in.PrevOut.Value)
		}

		for _, out := range tx.Outs {
			utxo[string(out.Pubkey)] += int64(out.Value)
		}
	}
	for addr, val := range utxo {
		if int64(service.utxo[addr])+val < 0 {
			return errors.ErrAccountNotEnoughValues
		}
	}
	for addr, val := range utxo {
		service.utxo[addr] += uint64(val)
		if service.utxo[addr] == 0 {
			delete(service.utxo, addr)
		}
	}
	return nil
}
