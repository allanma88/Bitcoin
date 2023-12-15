package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
	"encoding/json"
	"fmt"
	"os"
)

const (
	MEMPOOL = "mempool"
)

type MemPool struct {
	maxTxSize int
	mempool   *collection.SortedSet[string, uint64, *model.Transaction]
}

func NewMemPool(maxTxSize int) *MemPool {
	return &MemPool{
		maxTxSize: maxTxSize,
		mempool:   collection.NewSortedSet[string, uint64, *model.Transaction](),
	}
}

func (pool *MemPool) Get(hash []byte) *model.Transaction {
	return pool.mempool.Get1(string(hash))
}

func (pool *MemPool) Put(tx *model.Transaction) {
	pool.mempool.Insert(string(tx.Hash), tx.Fee, tx)
	if pool.mempool.Len() > pool.maxTxSize {
		min := pool.mempool.Min()
		pool.mempool.Remove(string(min.Hash), min.Fee)
	}
}

func (pool *MemPool) TopMax(n int) []*model.Transaction {
	return pool.mempool.TopMax(0, n)
}

func (pool *MemPool) Remove(txs []*model.Transaction) {
	for _, tx := range txs {
		pool.mempool.Remove(string(tx.Hash), tx.Fee)
	}
}

func (pool *MemPool) Load(dir string) error {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, MEMPOOL))
	if err != nil {
		return err
	}

	var txs []*model.Transaction
	if err := json.Unmarshal(data, &txs); err != nil {
		return err
	}

	for _, tx := range txs {
		pool.Put(tx)
	}
	return nil
}

func (pool *MemPool) Save(dir string) error {
	txs := pool.mempool.TopMax(0, pool.mempool.Len())
	data, err := json.Marshal(txs)
	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", dir, MEMPOOL))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}
