package service

import (
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	UTXO = "utxo"
)

type UtxoService struct {
	utxo      map[string]uint64
	timestamp time.Time
}

func (service *UtxoService) MarshalJSON() ([]byte, error) {
	return json.Marshal(service)
}

func (service *UtxoService) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, service)
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

	if len(blocks) > 0 {
		service.timestamp = blocks[len(blocks)-1].Time
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

func (service *UtxoService) Load(dir string) error {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, UTXO))
	if err != nil {
		return err
	}

	var s UtxoService
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	service.timestamp = s.timestamp
	for addr, val := range s.utxo {
		service.utxo[addr] = val
	}
	return nil
}

func (service *UtxoService) Save(dir string) error {
	// TODO: lock?
	data, err := json.Marshal(service)
	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", dir, UTXO))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}
