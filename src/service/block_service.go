package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/infra"
	"Bitcoin/src/model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	MaxBlocksPerGetBlockReq = 100
	Genesis                 = "Genesis"
)

//TODO: more test cases

type BlockService struct {
	database.IBlockDB
}

func NewBlockService(blockDB database.IBlockDB) *BlockService {
	return &BlockService{
		IBlockDB: blockDB,
	}
}

func (s *BlockService) GetBlocks(lastBlockHash []byte, blockhashes [][]byte, ctxs []context.Context) ([]*model.Block, uint64, error) {
	for _, blockHash := range blockhashes {
		ancestors, err := s.ancestors(lastBlockHash, blockHash, ctxs)
		if err != nil {
			return nil, 0, err
		}

		if ancestors != nil {
			var end uint64
			if len(ancestors) > 0 {
				end = ancestors[len(ancestors)-1].Number
			}
			return ancestors[:MaxBlocksPerGetBlockReq], end, nil
		}
	}
	return nil, 0, nil
}

func (s *BlockService) Validate(block *model.Block) error {
	hash, err := validateHash[*model.Block](block.Hash, block)
	if err != nil {
		return err
	}

	err = validateTimestamp(block.Time)
	if err != nil {
		return err
	}

	existBlock, err := s.GetBlock(hash, false)
	if err != nil {
		return err
	}
	if existBlock != nil {
		return errors.ErrBlockExist
	}

	prevBlock, err := s.GetBlock(block.Prevhash, false)
	if err != nil {
		return err
	}
	if prevBlock == nil {
		return errors.ErrPrevBlockNotFound
	}
	if block.Number != prevBlock.Number+1 {
		return errors.ErrBlockNumberInvalid
	}
	if prevBlock.Time.Compare(block.Time) > 0 {
		return errors.ErrBlockTooLate
	}

	err = validateDifficulty(block.Hash, block.Difficulty)
	if err != nil {
		return err
	}

	err = validateRootHash(block.RootHash, block.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *BlockService) GetBlocksOfChain(applyChain, rollbackChain *model.Chain) ([]*model.Block, []*model.Block, error) {
	applyBlocks := make([]*model.Block, 0)
	rollbackBlocks := make([]*model.Block, 0)

	applyBlock, err := s.GetBlock(applyChain.LastBlockHash, true)
	if err != nil {
		return nil, nil, err
	}
	rollbackBlock, err := s.GetBlock(rollbackChain.LastBlockHash, true)
	if err != nil {
		return nil, nil, err
	}

	for !bytes.Equal(rollbackBlock.Hash, applyBlock.Hash) {
		rollbackBlocks = append(rollbackBlocks, rollbackBlock)
		applyBlocks = append(applyBlocks, applyBlock)

		applyBlock, err = s.GetBlock(applyBlock.Prevhash, true)
		if err != nil {
			return nil, nil, err
		}
		rollbackBlock, err = s.GetBlock(rollbackBlock.Prevhash, true)
		if err != nil {
			return nil, nil, err
		}
	}
	return applyBlocks, rollbackBlocks, nil
}

func (s *BlockService) TryAddGenesis(dir string, level uint64) error {
	size, err := s.Size()
	if err != nil {
		return err
	}

	if size == 0 {
		data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, Genesis))
		if err != nil {
			return err
		}

		var outs []*model.Out
		if err := json.Unmarshal(data, &outs); err != nil {
			return err
		}

		block, err := model.MakeGenesisBlock(outs, level)
		if err != nil {
			return err
		}

		if err = s.SaveBlock(block); err != nil {
			return err
		}
	}

	return nil
}

func validateDifficulty(hash []byte, difficulty float64) error {
	actual := infra.ComputeDifficulty(hash)
	if actual > difficulty {
		return errors.ErrBlockNonceInvalid
	}
	return nil
}

func validateRootHash(roothash []byte, tree *collection.MerkleTree[*model.Transaction]) error {
	valid, err := tree.Validate()
	if err != nil {
		return err
	}

	if !valid {
		log.Printf("content is invalid")
		return errors.ErrBlockContentInvalid
	}

	if !bytes.Equal(roothash, tree.Table[len(tree.Table)-1][0].Hash) {
		log.Printf("content hash mismatch with root hash")
		return errors.ErrBlockContentInvalid
	}

	return nil
}

func (s *BlockService) ancestors(lastBlockHash, ancestor []byte, ctxs []context.Context) ([]*model.Block, error) {
	ancestors := make([]*model.Block, 0)
	for !bytes.Equal(ancestor, lastBlockHash) {
		for _, ctx := range ctxs {
			err := context.Cause(ctx)
			if err != nil {
				return nil, err
			}
		}

		//TODO: split the GetBlock api to GetBlockHeader and GetBlockContent, not include body here
		block, err := s.GetBlock(lastBlockHash, true)
		if err != nil {
			return nil, err
		}
		ancestors = append([]*model.Block{block}, ancestors...)
		lastBlockHash = block.Prevhash
	}
	return nil, nil
}
