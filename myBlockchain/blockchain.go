package main

import "fmt"

// blockchain 表示了一条区块链
type blockchain struct {
	blocks []*block
}

func newBlockchain() *blockchain {
	return &blockchain{
		// 创世区块作为区块链的第一个区块
		blocks: []*block{makeGenesisBlock()},
	}
}

// addBlock 往 bc 中添加新的区块
func (bc *blockchain) addBlock(data string) {
	preBlockHash := bc.blocks[len(bc.blocks)-1].hash
	b := newBlock(data, preBlockHash)
	bc.blocks = append(bc.blocks, b)
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
