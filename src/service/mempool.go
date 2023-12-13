package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
)

type MemPool struct {
	maxTxSize int
	mempool   *collection.SortedSet[*model.Transaction]
	txmap     map[string]*model.Transaction
}

func NewMemPool(maxTxSize int) *MemPool {
	return &MemPool{
		maxTxSize: maxTxSize,
		mempool:   collection.NewSortedSet[*model.Transaction](),
		txmap:     make(map[string]*model.Transaction),
	}
}

func (service *MemPool) Get(hash []byte) *model.Transaction {
	return service.txmap[string(hash)]
}

func (service *MemPool) Put(tx *model.Transaction) {
	service.mempool.Insert(tx)
	service.txmap[string(tx.Hash)] = tx
	if service.mempool.Len() > service.maxTxSize {
		min := service.mempool.Min()
		service.mempool.Remove(min)
		delete(service.txmap, string(min.Hash))
	}
}

func (service *MemPool) TopMax(n int) []*model.Transaction {
	return service.mempool.TopMax(0, n)
}

func (service *MemPool) Remove(txs []*model.Transaction) {
	for _, tx := range txs {
		service.mempool.Remove(tx)
		delete(service.txmap, string(tx.Hash))
	}
}

func (service *TransactionService) Save() {
}
