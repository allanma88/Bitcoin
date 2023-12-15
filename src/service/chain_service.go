package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
	"sync"
)

//TODO: test cases
//TODO: remove too old branches

type ChainService struct {
	chains *collection.SortedSet[string, uint64, *model.Chain]
	lock   sync.Mutex
}

func NewChainService() *ChainService {
	return &ChainService{
		chains: collection.NewSortedSet[string, uint64, *model.Chain](),
		lock:   sync.Mutex{},
	}
}

func (s *ChainService) GetMainChain() *model.Chain {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.chains.Max()
}

func (s *ChainService) GetChainHashes(m, n int) [][]byte {
	s.lock.Lock()
	defer s.lock.Unlock()

	blocks := s.chains.TopMax(m, n)
	blockHashes := make([][]byte, len(blocks))
	for i := 0; i < len(blocks); i++ {
		blockHashes[i] = blocks[i].LastBlockHash
	}
	return blockHashes
}

func (s *ChainService) SetChain(block *model.Block) (*model.Chain, *model.Chain) {
	s.lock.Lock()
	defer s.lock.Unlock()

	chain := s.chains.Get(string(block.Prevhash), block.Number-1)
	chain.LastBlockHash = block.Hash
	chain.Length = block.Number

	mainChain := s.chains.Max()
	if mainChain != chain {
		return chain, mainChain
	}

	return nil, nil
}

func (s *ChainService) ChainLen() int {
	return s.chains.Len()
}
