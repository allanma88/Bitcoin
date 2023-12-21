package service

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/database"
	"Bitcoin/src/errors"
	"Bitcoin/src/infra"
	"Bitcoin/src/model"
	"bytes"
	"log"
)

const (
	MaxBlocksPerGetBlockReq = 100
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

func (service *BlockService) GetBlocks(lastBlockHash []byte, blockhashes [][]byte) ([]*model.Block, uint64, error) {
	for _, blockHash := range blockhashes {
		ancestors, err := service.Ancestors(lastBlockHash, blockHash)
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

// TODO: make internal
func (service *BlockService) Ancestors(lastBlockHash, ancestor []byte) ([]*model.Block, error) {
	ancestors := make([]*model.Block, 0)
	for !bytes.Equal(ancestor, lastBlockHash) {
		//TODO: split the GetBlock api to GetBlockHeader and GetBlockContent, not include body here
		block, err := service.GetBlock(lastBlockHash, true)
		if err != nil {
			return nil, err
		}
		ancestors = append([]*model.Block{block}, ancestors...)
		lastBlockHash = block.Prevhash
	}
	return nil, nil
}

func (service *BlockService) Validate(block *model.Block) error {
	hash, err := validateHash[*model.Block](block.Hash, block)
	if err != nil {
		return err
	}

	err = validateTimestamp(block.Time)
	if err != nil {
		return err
	}

	existBlock, err := service.GetBlock(hash, false)
	if err != nil {
		return err
	}
	if existBlock != nil {
		return errors.ErrBlockExist
	}

	prevBlock, err := service.GetBlock(block.Prevhash, false)
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

func (service *BlockService) GetBlocksOfChain(applyChain, rollbackChain *model.Chain) ([]*model.Block, []*model.Block, error) {
	applyBlocks := make([]*model.Block, 0)
	rollbackBlocks := make([]*model.Block, 0)

	applyBlock, err := service.GetBlock(applyChain.LastBlockHash, true)
	if err != nil {
		return nil, nil, err
	}
	rollbackBlock, err := service.GetBlock(rollbackChain.LastBlockHash, true)
	if err != nil {
		return nil, nil, err
	}

	for !bytes.Equal(rollbackBlock.Hash, applyBlock.Hash) {
		rollbackBlocks = append(rollbackBlocks, rollbackBlock)
		applyBlocks = append(applyBlocks, applyBlock)

		applyBlock, err = service.GetBlock(applyBlock.Prevhash, true)
		if err != nil {
			return nil, nil, err
		}
		rollbackBlock, err = service.GetBlock(rollbackBlock.Prevhash, true)
		if err != nil {
			return nil, nil, err
		}
	}
	return applyBlocks, rollbackBlocks, nil
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
