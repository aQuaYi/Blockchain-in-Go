package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type block struct {
	timestamp    int64
	data         []byte
	preBlockHash []byte
	hash         []byte
	Nonce        int
}

func (b *block) String() string {
	pow := newProofOfWork(b)
	res := fmt.Sprintf("Prev. Hash: %x\n", b.preBlockHash)
	res += fmt.Sprintf("      Data: %s\n", b.data)
	res += fmt.Sprintf("       PoW: %t\n", pow.validate())
	res += fmt.Sprintf("This  Hash: %x\n", b.hash)
	return res
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

func (b *block) Serialize() []byte {
	var res bytes.Buffer

	enc := gob.NewEncoder(&res)
	err := enc.Encode(b)
	if err != nil {
		log.Fatal(err)
	}

	return res.Bytes()
}

// DeserializeBlock 把 kv 数据库中读取的数据，反向序列化成 *block 对象
func deserializeBlock(d []byte) *block {
	var block block

	dec := gob.NewDecoder(bytes.NewReader(d))
	err := dec.Decode(&block)
	if err != nil {
		log.Fatal(err)
	}

	return &block
}
