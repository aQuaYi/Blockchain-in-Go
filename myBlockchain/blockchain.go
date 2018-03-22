package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const (
	dbFile       = "bc.db"
	blocksBucket = "bc"
)

// blockchain 表示了一条区块链
type blockchain struct {
	tip []byte
	db  *bolt.DB
}

func newBlockchain() *blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := makeGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Fatal(err)
			}
			err = b.Put(genesis.hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.hash)
			tip = genesis.hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := blockchain{
		tip: tip,
		db:  db,
	}

	return &bc
}

// addBlock 往 bc 中添加新的区块
func (bc *blockchain) addBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	newBlock := newBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.hash)
		bc.tip = newBlock.hash

		return nil
	})
}

type blockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc blockchain) Iterator() *blockchainIterator {
	return &blockchainIterator{
		currentHash: bc.tip,
		db:          bc.db,
	}
}

func (bi *blockchainIterator) Next() *block {
	var block *block

	err := bi.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encBlock := b.Get(bi.currentHash)
		block = deserializeBlock(encBlock)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bi.currentHash = block.preBlockHash
	return block
}
