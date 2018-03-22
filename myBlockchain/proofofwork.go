package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const (
	// targetBits 控制 pow 的难度
	// targetBits 越小，范围的上限就越大，pow 就越容易
	targetBits = 20
	// maxNonce 防止 pow 时，溢出
	maxNonce = math.MaxInt64
)

// ProofOfWork 是工作量证明结构体
type ProofOfWork struct {
	block  *block
	target *big.Int
}

func newProofOfWork(b *block) *ProofOfWork {
	target := big.NewInt(1)
	// Lsh 是把 target 左移了
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
