package service

import (
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
)

const (
	UTXO = "utxo"
)

type UtxoService struct {
	utxo map[string]uint64
}

// TODO: test case: apply all or none
func (s *UtxoService) ApplyBalance(block *model.Block) error {
	utxo := make(map[string]int64)
	s.applyBalances(utxo, block)
	s.applyUtxo(utxo)
	return nil
}

// TODO: test case: apply all or none
func (s *UtxoService) SwitchBalances(rollbackBlocks, applyBlocks []*model.Block) error {
	//TODO: maybe merge rollback and apply
	utxo := make(map[string]int64)
	if err := s.rollbackBalances(utxo, rollbackBlocks); err != nil {
		return err
	}
	if err := s.applyBalances(utxo, applyBlocks...); err != nil {
		return err
	}

	s.applyUtxo(utxo)
	return nil
}

func (s *UtxoService) applyBalances(utxo map[string]int64, blocks ...*model.Block) error {
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
			if int64(s.utxo[addr])+val < 0 {
				return errors.ErrAccountNotEnoughValues
			}
		}
	}

	return nil
}

func (s *UtxoService) rollbackBalances(utxo map[string]int64, blocks []*model.Block) error {
	// utxo := make(map[string]int64)
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
			if int64(s.utxo[addr])+val < 0 {
				return errors.ErrAccountNotEnoughValues
			}
		}
	}

	return nil
}

func (s *UtxoService) applyUtxo(utxo map[string]int64) {
	for addr, val := range utxo {
		s.utxo[addr] += uint64(val)
		if s.utxo[addr] == 0 {
			delete(s.utxo, addr)
		}
	}
}
