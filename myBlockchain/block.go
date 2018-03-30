package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

// Block keeps block headers
type Block struct {
	Timestamp     int64          // 生成区块的时间戳
	Transactions  []*Transaction // 在区块中包含的交易
	PrevBlockHash []byte         // 前一个区块的哈希值
	Hash          []byte         // 这个区块的哈希值
	Nonce         int            // 这个区块的附加值，用于验证
}

// Serialize serializes the block
// 用于在 boltDB 中保存
// Serialize 把 b 以 gob 的方式转换成 []byte
func (b *Block) Serialize() []byte {
	// 存放结果，结果是 []byte
	// 结果有需要包含 io.Writer 接口
	// 所以采用 bytes.Buffer
	var result bytes.Buffer
	// encoder 负责把结果输出到 result
	encoder := gob.NewEncoder(&result)
	// 对 b 进行编码，输出到 result
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	// 输出生成的 []byte
	return result.Bytes()
}

// HashTransactions 返回区块中所有交易的哈希值
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	// 把所有交易内容的哈希值，按照顺序收集起来
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	// 利用总的内容重新生成哈希值
	// 得到全部交易的哈希值
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	// 根据参数生成新的区块
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	// 由区块生成新的 pow 对象
	pow := NewProofOfWork(block)
	// 进行 pow 运算
	// 得到了 block 在 nonce 的帮助下的，符合难度要求的哈希值
	nonce, hash := pow.Run()

	// 把 nonce 和 hash 存入 block 中
	block.Hash = hash[:]
	block.Nonce = nonce

	// 返回 block
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	// 生成创世区块
	// 创世区块只包含一个 coinbase 交易
	// 且
	// 创世区块的前一个区块的哈希值是 []byte{}
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// DeserializeBlock deserializes a block
// 从 boltDB 数据库取出区块的 gob 编码后，
// 由 DeserializeBlock 还原成 block 对象
func DeserializeBlock(d []byte) *Block {
	// 还原后的 Block 对象放在 block
	var block Block

	// bytes.NewReader(d) 让 d 具有了 io.Reader 接口
	decoder := gob.NewDecoder(bytes.NewReader(d))
	// decoder 会把还原好的 Block 对象放入 block 中
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	// 返回 block 变量的指针
	return &block
}
