package main

import "bytes"

// TXInput 是交易的输入值
type TXInput struct {
	// 假设此 input 所引用的 output 属于交易 tx
	Txid      []byte // tx.ID
	Vout      int    // NOTICE: output 在 tx.TXOutput 中的索引号
	Signature []byte // 签名
	PubKey    []byte // 签名时，所用的公钥
}

// UsesKey checks whether the address initiated the transaction
//
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
