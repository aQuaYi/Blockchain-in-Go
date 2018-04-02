package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	// nonce 所需尝试的最大值
	// 防止寻找 nonce 的 for 循环溢出
	maxNonce = math.MaxInt64
)

// 用于控制符合条件的区块链的哈希值范围
// targetBits 表示目标值左边有 targetBits 位，只能为 0
var targetBits = 20

// ProofOfWork represents a proof-of-work
// 工作量证明工作
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
// 着手准备进行工作量证明
// 以区块链指针作为输入变量
func NewProofOfWork(b *Block) *ProofOfWork {
	// target 就是哈希值的上限
	// 先生成变量
	target := big.NewInt(1)
	// 再进行位移操作
	target.Lsh(target, uint(256-targetBits))
	// 生成 pow 实例
	pow := &ProofOfWork{
		block:  b,
		target: target,
	}
	return pow
}

// 为进行下一次碰撞准备数据
// 由于 nonce 每次都会 +1
// 所以，每次都要重新准备
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	// 使用空切片，链接以下值
	data := bytes.Join(
		[][]byte{
			// 前一个区块的哈希值
			pow.block.PrevBlockHash,
			// 此区块，所有交易的哈希值
			pow.block.HashTransactions(),
			// 时间戳的哈希值
			IntToHex(pow.block.Timestamp),
			// 难度要求的哈希值
			IntToHex(int64(targetBits)),
			// 附加值的设置
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	// 此区块链的哈希值所代表的整数
	var hashInt big.Int
	// 此区块的哈希值
	var hash [32]byte
	// 附加值
	nonce := 0

	// 开始挖掘新带区块
	fmt.Printf("Mining a new block")
	// 寻找合适的 nonce 的 for 循环
	for nonce < maxNonce {
		// 利用新的 nonce 准备数据
		data := pow.prepareData(nonce)

		// 利用新数据生成哈希值
		hash = sha256.Sum256(data)
		// 显示输出哈希值
		fmt.Printf("\r%x", hash)
		// 把哈希值转换成 big.Int 型
		hashInt.SetBytes(hash[:])

		// 检验新哈希值是否符合要求
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	// 返回让哈希符合条件的 nonce 及其哈希值
	return nonce, hash[:]
}

// Validate validates block's PoW
// 验证区块链是否符合工作量证明的要求
func (pow *ProofOfWork) Validate() bool {
	// 哈希值所代表的整数值
	var hashInt big.Int

	// 按照 pow 的要求，准备 pow 所需要的数据
	data := pow.prepareData(pow.block.Nonce)
	//  生成哈希值
	hash := sha256.Sum256(data)
	//  查看哈希值所应对的整数值
	hashInt.SetBytes(hash[:])

	// 对比标准，得出结论
	isValid := hashInt.Cmp(pow.target) == -1

	// 返回结论
	return isValid
}
