package service

import (
	"Bitcoin/src/errors"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"log"
)

const (
	BatchSizePerGetBlocksReq = 10
)

type AddBlockFunc func(block *model.Block) error

type SyncService struct {
	chainService *ChainService
	nodeService  *NodeService
	addBlockFunc AddBlockFunc
}

func NewSyncService(chainService *ChainService, nodeService *NodeService, addBlockFunc AddBlockFunc) *SyncService {
	return &SyncService{
		chainService: chainService,
		nodeService:  nodeService,
		addBlockFunc: addBlockFunc,
	}
}

func (s *SyncService) SyncBlocks(addr string) {
	blockReqs, end, err := s.getBlocks(addr)
	for {
		if err != nil {
			log.Printf("get blocks from %v error: %v", addr, err)
			break
		}

		if blockReqs == nil && len(blockReqs) == 0 {
			break
		}

		var block *model.Block
		for _, blockReq := range blockReqs {
			block, err = model.BlockFrom(blockReq)
			if err != nil {
				break
			}
			err = s.addBlockFunc(block)
			if err == errors.ErrBlockExist {
				// it's possible that my main chain is not main chain anymore,
				// so I pull all blocks of the new main chain, the new main chain is side chain previously,
				// some blocks of the new main chain already exist, so ignore this error
				continue
			}
			if err != nil {
				break
			}
		}

		if err != nil {
			log.Printf("add block from %v error: %v", addr, err)
			break
		}

		if block.Number == end {
			break
		}

		blockReqs, end, err = s.nodeService.GetBlocks([][]byte{block.Hash}, addr)
	}
}

func (s *SyncService) getBlocks(addr string) ([]*protocol.BlockReq, uint64, error) {
	for lo := 0; lo < s.chainService.ChainLen(); lo += BatchSizePerGetBlocksReq {
		hi := lo + BatchSizePerGetBlocksReq
		if hi > s.chainService.ChainLen() {
			hi = s.chainService.ChainLen()
		}

		lastBlockHashes := s.chainService.GetChainHashes(lo, hi)
		blockReqs, end, err := s.nodeService.GetBlocks(lastBlockHashes, addr)
		if blockReqs != nil || err != nil {
			return blockReqs, end, err
		}
	}
	return nil, 0, nil
}
