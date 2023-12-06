package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
	"bytes"
	"sync"
)

//TODO: test cases

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

func (s *ChainService) GetChainHashes(m, n int) [][]byte {
	s.lock.Lock()
	defer s.lock.Unlock()

	blocks := s.chains.Top(m, n)
	blockHashes := make([][]byte, len(blocks))
	for i := 0; i < len(blocks); i++ {
		blockHashes[i] = blocks[i].Hash
	}
	return blockHashes
}

func (s *ChainService) SetChain(block *model.Block) ([]*model.Block, []*model.Block) {
	s.lock.Lock()
	defer s.lock.Unlock()

	lastBlock := s.chains.First()

	s.chains.Remove(block.PrevBlock)
	s.chains.Insert(block)

	applyBlocks := make([]*model.Block, 0)
	rollbackBlocks := make([]*model.Block, 0)

	if !bytes.Equal(lastBlock.Hash, block.Prevhash) {
		for !bytes.Equal(lastBlock.Hash, block.Hash) {
			rollbackBlocks = append(rollbackBlocks, lastBlock)
			applyBlocks = append(applyBlocks, block)

			block = block.PrevBlock
			lastBlock = lastBlock.PrevBlock
		}
		return applyBlocks, rollbackBlocks
	}

	return nil, nil
}

func (s *ChainService) ChainLen() int {
	return s.chains.Len()
}
