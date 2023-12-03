package service

import (
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"log"
)

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

func (service *UtxoService) UpdateBalances(txs []*model.Transaction) {
	for _, tx := range txs {
		log.Printf("updating balance for %x", tx.Hash)
		for _, in := range tx.Ins {
			service.reduceBalance(in.PrevOut)
		}

		for _, out := range tx.Outs {
			service.addBalance(out)
		}
		log.Printf("updated balance for %x", tx.Hash)
	}
}

func (service *UtxoService) addBalance(out *model.Out) {
	log.Printf("add %x to uxto", out.Pubkey[:10])
	service.utxo[string(out.Pubkey)] += out.Value
}

func (service *UtxoService) reduceBalance(out *model.Out) error {
	if service.utxo == nil {
		log.Fatal("uxto is nil")
	}
	if out == nil {
		log.Fatal("out is nil")
	}
	if service.utxo[string(out.Pubkey)] < out.Value {
		return errors.ErrAccountNotEnoughValues
	}

	service.utxo[string(out.Pubkey)] -= out.Value
	if service.utxo[string(out.Pubkey)] == 0 {
		log.Printf("remove %x from uxto", out.Pubkey[:10])
		delete(service.utxo, string(out.Pubkey))
	}
	return nil
}
