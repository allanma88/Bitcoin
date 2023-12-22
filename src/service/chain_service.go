package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

const (
	Stat = "stat"
)

//TODO: test cases
//TODO: remove too old branches

type ChainService struct {
	*UtxoService
	chains *collection.SortedSet[string, uint64, *model.Chain]
	lock   sync.Mutex
}

func NewChainService(utxo map[string]uint64) *ChainService {
	utxoService := &UtxoService{
		utxo: utxo,
	}
	return &ChainService{
		UtxoService: utxoService,
		chains:      collection.NewSortedSet[string, uint64, *model.Chain](),
		lock:        sync.Mutex{},
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

func (s *ChainService) ApplyChain(block *model.Block) (*model.Chain, *model.Chain) {
	s.lock.Lock()
	defer s.lock.Unlock()

	chain := s.chains.Get(string(block.Prevhash), block.Number-1)
	if chain != nil {
		chain.LastBlockHash = block.Hash
		chain.Length = block.Number
	} else {
		s.chains.Insert(string(chain.LastBlockHash), chain.Length, chain)
	}

	mainChain := s.chains.Max()
	if mainChain != chain {
		return chain, mainChain
	}

	return nil, nil
}

func (s *ChainService) ChainLen() int {
	return s.chains.Len()
}

type snapshot struct {
	Chains []*model.Chain
	Utxo   map[string]uint64
}

func (s *ChainService) Load(dir string) ([]*model.Chain, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, Stat))
	if errors.Is(err, fs.ErrNotExist) {
		return []*model.Chain{}, nil
	}
	if err != nil {
		return nil, err
	}

	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}

	for _, chain := range snap.Chains {
		s.chains.Insert(string(chain.LastBlockHash), chain.Length, chain)
	}
	s.utxo = snap.Utxo
	return snap.Chains, nil
}

func (s *ChainService) Save(dir string) error {
	snap := snapshot{
		Chains: s.chains.TopMax(0, s.ChainLen()),
		Utxo:   s.utxo,
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", dir, Stat))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}
