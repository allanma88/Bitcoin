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

func (service *UtxoService) AddBalance(out *model.Out) {
	log.Printf("add %x to uxto", out.Pubkey[:10])
	service.utxo[string(out.Pubkey)] += out.Value
}

func (service *UtxoService) ReduceBalance(out *model.Out) error {
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
