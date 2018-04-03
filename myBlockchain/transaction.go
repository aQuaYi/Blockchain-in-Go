package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction 是交易
type Transaction struct {
	// ID 是
	// 当 Transaction.ID 的内容为 nil 时
	// 由 Transaction 包含的其他内容的 gob 编码生成的哈希值 // 详见 SetID 方法
	// 所以，Transaction 的 ID 其实是 哈希值
	ID []byte
	// Vin 要完成此交易，所有引用的输入的集合
	Vin []TXInput
	// Vout 完成此交易后，所有的产生的输出的集合
	Vout []TXOutput
}

// IsCoinbase 返回 true 如果 tx 是一个 coinbase 交易
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && // coinbase 只引用了一个输入
		len(tx.Vin[0].Txid) == 0 && // 这唯一的输入，所引用的输出所在的区块的 ID 是空的
		tx.Vin[0].Vout == -1 // 这唯一的输入，所引用的输出的索引号为 -1
}

// SetID 为此 transaction 设置 ID
// ID 是根据交易中输入输出的内容生成的哈希值
func (tx *Transaction) SetID() {
	// encoded 是 tx 的序列化编码
	var encoded bytes.Buffer
	// hash 是序列化编码提供生成的哈希值
	var hash [32]byte

	// 先进行序列化工作
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	// 再由序列化的值，生成哈希值
	hash = sha256.Sum256(encoded.Bytes())

	// 最后，设置为 tx 的 ID
	tx.ID = hash[:]
}

// TXOutput 是交易的输出值
type TXOutput struct {
	Value        int    // 此 output 的数值
	ScriptPubkey string // 被 input 引用时，用于验证引用者是否具有所有权
}

// CanUnlockOutputWith 返回 true，如果 unlockingData 可以解锁此 TXInput
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith 返回 true，如果 unlockingData 可以解锁 out 的话
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubkey == unlockingData
}

// NewCoinbaseTX 创建 coinbase 交易
// 就是没有输入的交易
func NewCoinbaseTX(to, data string) *Transaction {
	// 为输入准备数据
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// 生成输入
	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}

	// 生成输出
	txout := TXOutput{
		Value:        subsidy,
		ScriptPubkey: to,
	}

	// 生成交易
	tx := Transaction{
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}
	// 设置交易 ID
	tx.SetID()

	return &tx
}

// NewUTXOTransaction 会创建一个 transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	// 为交易所需的 inputs 和 outputs 创建变量
	var inputs []TXInput
	var outputs []TXOutput

	// 在区块链中搜寻所有属于 from 的 coin
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	// 所有属于 from 的 coin 数量不足此次交易的数量
	// 无法生成此次交易
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// acc >= amount，说明 from 有足够的 coin 完成此次交易
	// Build a list of inputs
	for txid, outs := range validOutputs {
		// 获取可引用输出所在的交易的 ID
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		// 获取可引用输出的数量
		for _, out := range outs {
			// 利用 txID 和 out 一起生成
			// 作为新交易的 input
			input := TXInput{
				Txid:      txID,
				Vout:      out,
				ScriptSig: from,
			}
			// 把所有的新 input 收集到 inputs 中
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	// 生成此次交易的主要输出
	outputs = append(outputs, TXOutput{
		Value:        amount,
		ScriptPubkey: to,
	})
	// acc > amount 的时候
	// 需要找零给 from
	// 所以，还需要一个输出给 from
	if acc > amount {
		outputs = append(outputs, TXOutput{
			Value:        acc - amount,
			ScriptPubkey: from,
		})
	}

	// 真正生成交易
	tx := Transaction{
		Vin:  inputs,
		Vout: outputs,
	}
	tx.SetID()

	return &tx
}

// Sign signs each input of a Transaction
// 对交易的每一个输入进行签名
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	// 确保每一个输入的所引用的输出所在的交易，都存在
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	// 对交易中包含的信息进行裁剪，只留下需要签名的部分。
	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}
