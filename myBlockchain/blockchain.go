package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(nil, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
}

// Iterator ...
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// NewBlockchain 使用 genesis Block 创建一条新的区块
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("没有找到区块链数据库。请先创建一个")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bc := Blockchain{
		tip: tip,
		db:  db,
	}

	return &bc
}

// CreateBlockchain 创建一个新的区块链数据库文件
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("区块链已经存在")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Fatal(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Fatal(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Fatal(err)
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	bc := Blockchain{
		tip: tip,
		db:  db,
	}

	return &bc
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
// TODO: 弄清楚这个方法
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	// 存放有没有被引用的输出的所有交易
	var unspentTXs []Transaction
	// 存放所有已经被引用的输出
	spentTXOs := make(map[string][]int)
	// 区块链的迭代器
	bci := bc.Iterator()

	for {
		// 从最新的区块开始迭代
		block := bci.Next()

		// 遍历此 block 的所有交易
		for _, tx := range block.Transaction {
			// TODO: txID 到底是什么鬼
			// 获取这个交易的 ID
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			// 遍历此交易的所有输出
			for outIdx, out := range tx.Vout {
				// 如果 spentTXOs 中存在 txID 的记录
				if spentTXOs[txID] != nil {
					// 遍历此 txID 的所有记录
					for _, spentOut := range spentTXOs[txID] {
						// 如果存在一样的索引号
						// 则跳过这个交易
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// 如果输出 out 可以被 address 解锁
				// 说明此 out 是address 还没有花的钱
				if out.CanBeUnlockedWith(address) {
					// 把这个交易放入 unspentTXs
					unspentTXs = append(unspentTXs, *tx)
				}

			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindSpendableOutputs 寻找并返回没有花掉的输出
// TODO: 弄清楚这个方法
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
