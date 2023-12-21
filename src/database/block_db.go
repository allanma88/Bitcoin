package database

import (
	"Bitcoin/src/collection"
	"Bitcoin/src/model"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	BlockTable        = "Block"
	BlockIndexTable   = "BlockIndex"
	BlockContentTable = "BlockContent"
	TxTable           = "Transaction"
)

type IBlockDB interface {
	SaveBlock(block *model.Block) error
	GetBlock(hash []byte, includeBody bool) (*model.Block, error)
	FilterBlock(prevBlockHash []byte) ([]*model.Block, error)
	SaveTx(tx *model.Transaction) error
	GetTx(hash []byte) (*model.Transaction, error)
	Close() error
}

type BlockDB struct {
	IBaseDB
}

func NewBlockDB(db *leveldb.DB) IBlockDB {
	basedb := &BaseDB{Database: db}
	blockdb := &BlockDB{IBaseDB: basedb}
	return blockdb
}

func (db *BlockDB) SaveBlock(block *model.Block) error {
	batch := db.StartBatch()

	if err := batch.Save([]byte(BlockTable), block.Hash, block); err != nil {
		return err
	}

	if err := batch.Save([]byte(BlockIndexTable), block.Prevhash, block.Hash); err != nil {
		return err
	}

	if err := batch.Save([]byte(BlockContentTable), block.RootHash, block.Body); err != nil {
		return err
	}

	for _, tx := range block.Body.GetVals() {
		if err := batch.Save([]byte(TxTable), tx.Hash, tx); err != nil {
			return err
		}
	}

	return db.EndBatch(batch)
}

func (db *BlockDB) GetBlock(hash []byte, includeBody bool) (*model.Block, error) {
	var block model.Block
	has, err := db.Get([]byte(BlockTable), hash, &block)
	if !has || err != nil {
		return nil, err
	}

	if includeBody {
		var content collection.MerkleTree[*model.Transaction]
		has, err = db.Get([]byte(BlockContentTable), block.RootHash, &content)
		if !has || err != nil {
			return nil, err
		}

		for i := 0; i < len(content.Table[0]); i++ {
			txhash := content.Table[0][i].Hash

			tx, err := db.GetTx(txhash)
			if err != nil {
				return nil, err
			}
			content.Table[0][i].Val = tx
		}
		block.Body = &content
	}

	return &block, nil
}

func (db *BlockDB) FilterBlock(prevBlockHash []byte) ([]*model.Block, error) {
	datalist, err := db.Filter([]byte(BlockIndexTable), prevBlockHash)
	if err != nil {
		return nil, err
	}

	blocks := make([]*model.Block, len(datalist))
	for i, data := range datalist {
		var block model.Block
		if err = json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		blocks[i] = &block
	}

	return blocks, nil
}

func (db *BlockDB) SaveTx(tx *model.Transaction) error {
	return db.Save([]byte(TxTable), tx.Hash, tx)
}

func (db *BlockDB) GetTx(hash []byte) (*model.Transaction, error) {
	var tx model.Transaction
	has, err := db.Get([]byte(TxTable), hash, &tx)
	if !has || err != nil {
		return nil, err
	}

	return &tx, nil
}
