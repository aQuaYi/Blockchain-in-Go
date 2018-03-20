package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"
)

const targetBits = 20
const maxNonce = math.MaxInt64

type block struct {
	timestamp    int64
	data         []byte
	preBlockHash []byte
	hash         []byte
	Nonce        int
}

func newBlock(data string, preBlockHash []byte) *block {
	b := &block{
		timestamp:    time.Now().Unix(),
		preBlockHash: preBlockHash,
		data:         []byte(data),
	}

	pow := newProofOfWork(b)
	nonce, hash := pow.Run()

	b.hash = hash
	b.Nonce = nonce

	return b
}

// blockchain 表示了一条区块链
type blockchain struct {
	blocks []*block
}

// addBlock 往 bc 中添加新的区块
func (bc *blockchain) addBlock(data string) {
	preBlockHash := bc.blocks[len(bc.blocks)-1].hash
	b := newBlock(data, preBlockHash)
	bc.blocks = append(bc.blocks, b)
}

func newBlockchain() *blockchain {
	return &blockchain{
		blocks: []*block{makeGenesisBlock()},
	}
}

func makeGenesisBlock() *block {
	return newBlock("Genesis Block", []byte{})
}

func (bc *blockchain) print() {
	for _, b := range bc.blocks {
		fmt.Printf("pre  Hash: %x\n", b.preBlockHash)
		fmt.Printf("     data: %s\n", b.data)
		pow := newProofOfWork(b)
		fmt.Printf("      PoW: %t\n", pow.validate())
		fmt.Printf("this Hash: %x\n", b.hash)
		fmt.Println()
	}
}

// ProofOfWork 是工作量证明结构体
type ProofOfWork struct {
	block  *block
	target *big.Int
}

func newProofOfWork(b *block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		block:  b,
		target: target,
	}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.preBlockHash,
			pow.block.data,
			IntToHex(pow.block.timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// IntToHex converts an int64 to a byte array
// TODO: 解释这个函数
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// Run 是进行工作
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		}

		nonce++
	}

	fmt.Printf("\n\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}

func main() {
	bc := newBlockchain()

	bc.addBlock("Send 1 BTC to Alice")
	bc.addBlock("Send 2 BTC to Bob")
	bc.addBlock("Send 3 BTC to Candy")

	bc.print()
}
