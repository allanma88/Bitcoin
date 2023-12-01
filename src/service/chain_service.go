package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
	"bytes"
	"sync"
)

type ChainService struct {
	chains *collection.SortedSet[*model.Block]
	lock   sync.Mutex
}

func NewChainService() *ChainService {
	return &ChainService{
		chains: collection.NewSortedSet[*model.Block](),
		lock:   sync.Mutex{},
	}
}

func (s *ChainService) GetMainChain() *model.Block {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.chains.First()
}

func (s *ChainService) GetChainHashes(n int) [][]byte {
	s.lock.Lock()
	defer s.lock.Unlock()

	blocks := s.chains.Top(n)
	blockHashes := make([][]byte, len(blocks))
	for i := 0; i < len(blocks); i++ {
		blockHashes[i] = blocks[i].Hash
	}
	return blockHashes
}

func (s *ChainService) SetChain(block *model.Block) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	lastBlock := s.chains.First()
	isMainChain := bytes.Equal(lastBlock.Hash, block.Prevhash)

	s.chains.Remove(block.PrevBlock)
	s.chains.Insert(block)

	return isMainChain
}
