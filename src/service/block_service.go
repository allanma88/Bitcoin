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
	FetchBatch              = 100
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

func (service *BlockService) GetBlocks(mainChain *model.Block, blockhashes [][]byte) ([]*model.Block, uint64, error) {
	for _, blockHash := range blockhashes {
		block, err := service.GetBlock(blockHash)
		if err != nil {
			return nil, 0, err
		}
		ancestors := mainChain.Ancestors(block)
		if ancestors != nil {
			var end uint64
			if len(ancestors) > 0 {
				end = ancestors[len(ancestors)-1].Number
			} else {
				end = block.Number
			}
			return ancestors[:MaxBlocksPerGetBlockReq], end, nil
		}
	}
	return nil, 0, nil
}

// func (service *BlockService) SaveBlock(block *model.Block) error {
// 	return service.SaveBlock(block)
// }

func (service *BlockService) LoadBlocks(utxo *UtxoService) error {
	for {
		blocks, err := service.FilterBlock(utxo.timestamp, FetchBatch)
		if err != nil {
			return err
		}
		if err = utxo.ApplyBalances(blocks...); err != nil {
			return err
		}
		if len(blocks) < FetchBatch {
			break
		}
	}

	return nil
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

	existBlock, err := service.GetBlock(hash)
	if err != nil {
		return err
	}
	if existBlock != nil {
		return errors.ErrBlockExist
	}

	prevBlock, err := service.GetBlock(block.Prevhash)
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
	block.PrevBlock = prevBlock

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
