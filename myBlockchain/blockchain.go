package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

// 区块链数据库的文件名称
const dbFile = "blockchain.db"

// 区块在 boltDB 中 Bucket 的名称
const blocksBucket = "blocks"

// 创世区块所包含的消息
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain keeps a sequence of Blocks
// 区块链结构体
type Blockchain struct {
	tip []byte   // 最新的区块的哈希值
	db  *bolt.DB // 存放区块的数据库文件
}

// BlockchainIterator is used to iterate over blockchain blocks
// 区块链迭代器，用于依次访问从最新到最旧的全部区块
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// MineBlock mines a new block with the provided transactions
// 挖掘
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

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
		for _, tx := range block.Transactions {
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
						// TODO: 此处的代码
						// if spentOut == out.Value {
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
			// 如果 tx 不是 Coinbase 交易的话
			// tx 一定是对别的交易的输出进行了引用
			// 要把这些交易找出来，放入 spentTXOs
			if tx.IsCoinbase() == false {
				// 对于此交易中的所有的 input
				for _, in := range tx.Vin {
					// 如果 in 能由 address 生成
					if in.CanUnlockOutputWith(address) {
						// 获取 input 所引用的 output 所在交易的 ID
						inTxID := hex.EncodeToString(in.Txid)
						// TODO: 此处的代码
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			// 已经遍历完成了所有区块
			// 结束循环
			break
		}
	}

	// 返回所有没有花费的交易
	return unspentTXs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	// 收集所有未花费的输出
	var UTXOs []TXOutput
	// 收集所有含有 address 的未花费输出的交易
	unspentTransactions := bc.FindUnspentTransactions(address)

	// 遍历所有含有 address 的未花费输出的交易
	for _, tx := range unspentTransactions {
		// 遍历交易中的所有输出
		for _, out := range tx.Vout {
			// 如果输出能够被 address 解锁 → 这是 address 的未花费的输出
			if out.CanBeUnlockedWith(address) {
				//  把输出 out 放入 UTXOs 中
				UTXOs = append(UTXOs, out)
			}
		}
	}

	// 返回所有找到的 address 的未花费的输出
	return UTXOs
}

// FindSpendableOutputs 寻找并返回没有花掉的输出
// TODO: 弄清楚这个方法
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	// 没有被引用的输出们
	unspentOutputs := make(map[string][]int)
	// 包含 address 可以引用的输出的交易们
	unspentTXs := bc.FindUnspentTransactions(address)
	// 所有未引用输出的累计数量
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		// 获取此交易的 ID
		txID := hex.EncodeToString(tx.ID)
		// 对于此交易的每一个输出而言
		for outIdx, out := range tx.Vout {
			// 如果输出能被 address 解锁 → 说明，这是 address 的钱
			// 且，还没有累计到所需的数量
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				// 那就算上这个输出吧
				accumulated += out.Value
				// 并把这个交易带上
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				// TODO: 核查此处代码
				// unspentOutputs[txID] = append(unspentOutputs[txID], out.Value)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	// 返回已经累计的钱数 和 包含这些钱的输出们
	return accumulated, unspentOutputs
}

// Iterator 返回区块链的迭代器
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
// TODO: 暂时 address 参数没用
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
		log.Panic(err)
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
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{
		tip: tip,
		db:  db,
	}

	return &bc
}
