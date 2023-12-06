package service

import (
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
)

type UtxoService struct {
	utxo map[string]uint64
}

func NewUtxoService() *UtxoService {
	return &UtxoService{utxo: make(map[string]uint64)}
}

// TODO: test case
func (service *UtxoService) SwitchBalances(rollbackBlocks, applyBlocks []*model.Block) error {
	//TODO: maybe merge rollback and apply
	if err := service.rollbackBalances(rollbackBlocks); err != nil {
		return err
	}
	if err := service.ApplyBalances(applyBlocks...); err != nil {
		return err
	}
	return nil
}

// TODO: test case: apply all or none
func (service *UtxoService) ApplyBalances(blocks ...*model.Block) error {
	utxo := make(map[string]int64)
	for _, block := range blocks {
		for _, tx := range block.GetTxs() {
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
func (service *UtxoService) rollbackBalances(blocks []*model.Block) error {
	utxo := make(map[string]int64)
	for _, block := range blocks {
		for _, tx := range block.GetTxs() {
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
	}

	for addr, val := range utxo {
		service.utxo[addr] += uint64(val)
		if service.utxo[addr] == 0 {
			delete(service.utxo, addr)
		}
	}
	return nil
}
