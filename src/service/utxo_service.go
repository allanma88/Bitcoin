package service

import (
	"Bitcoin/src/model"
	"log"
)

const (
	UTXO = "utxo"
)

type UtxoService struct {
	utxo map[string]uint64
}

func (s *UtxoService) ApplyTx(tx *model.Transaction) {
	utxo := make(map[string]int64)
	s.applyTx(utxo, tx)
	s.applyUtxo(utxo)
}

// TODO: test case: apply all or none
func (s *UtxoService) ApplyBlock(block *model.Block) {
	utxo := make(map[string]int64)
	s.applyBlocks(utxo, block)
	s.applyUtxo(utxo)
}

// TODO: test case: apply all or none
func (s *UtxoService) SwitchBlocks(rollbackBlocks, applyBlocks []*model.Block) {
	utxo := make(map[string]int64)
	s.rollbackBlocks(utxo, rollbackBlocks)
	s.applyBlocks(utxo, applyBlocks...)
	s.applyUtxo(utxo)
}

func (s *UtxoService) applyTx(utxo map[string]int64, tx *model.Transaction) {
	for _, in := range tx.Ins {
		utxo[string(in.PrevOut.Pubkey)] -= int64(in.PrevOut.Value)
	}

	for _, out := range tx.Outs {
		utxo[string(out.Pubkey)] += int64(out.Value)
	}
}

func (s *UtxoService) applyBlocks(utxo map[string]int64, blocks ...*model.Block) {
	for _, block := range blocks {
		for _, tx := range block.GetTxs() {
			s.applyTx(utxo, tx)
		}
	}
}

func (s *UtxoService) rollbackBlocks(utxo map[string]int64, blocks []*model.Block) {
	for _, block := range blocks {
		for _, tx := range block.GetTxs() {
			for _, in := range tx.Ins {
				utxo[string(in.PrevOut.Pubkey)] += int64(in.PrevOut.Value)
			}

			for _, out := range tx.Outs {
				utxo[string(out.Pubkey)] -= int64(out.Value)
			}
		}
	}
}

func (s *UtxoService) applyUtxo(utxo map[string]int64) {
	for addr, val := range utxo {
		if int64(s.utxo[addr])+val < 0 {
			log.Fatalf("not enough values for %v, only %d, but need substract %d", addr, s.utxo[addr], val)
		}
		s.utxo[addr] += uint64(val)
		if s.utxo[addr] == 0 {
			delete(s.utxo, addr)
		}
	}
}
