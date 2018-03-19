package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

type block struct {
	timestamp    int64
	preBlockHash []byte
	data         []byte
	hash         []byte
}

func (b *block) setHash() {
	timestamp := []byte(strconv.FormatInt(b.timestamp, 10))
	header := bytes.Join([][]byte{timestamp, b.preBlockHash, b.data}, []byte{})

	hash := sha256.Sum256(header)

	b.hash = hash[:]
}

func newBlock(data string, preBlockHash []byte) *block {
	b := &block{
		timestamp:    time.Now().Unix(),
		preBlockHash: preBlockHash,
		data:         []byte(data),
	}

	b.setHash()

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
		fmt.Printf("this Hash: %x\n", b.hash)
		fmt.Println()
	}
}

func main() {
	bc := newBlockchain()

	bc.addBlock("Send 1 BTC to Alice")
	bc.addBlock("Send 2 BTC to Bob")

	bc.print()
}
