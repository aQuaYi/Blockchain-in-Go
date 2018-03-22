package main

import "time"

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
func makeGenesisBlock() *block {
	// 创世区块的 preBlockhash 为空
	return newBlock("Genesis Block", []byte{})
}
