package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction 是交易
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// IsCoinbase 返回 true 如果 tx 是一个 coinbase 交易
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 &&
		len(tx.Vin[0].Txid) == 0 &&
		tx.Vin[0].Vout == -1
}

// SetID 为此 transaction 设置 ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Fatal(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// TXInput 是交易jk的输入值
type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

// TXOutput 是交易的输出值
type TXOutput struct {
	Value        int
	ScriptPubkey string
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
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}

	txout := TXOutput{
		Value:        subsidy,
		ScriptPubkey: to,
	}

	tx := Transaction{
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction 会创建一个 transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Fatal(err)
		}

		for _, out := range outs {
			input := TXInput{
				Txid:      txID,
				Vout:      out,
				ScriptSig: from,
			}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, TXOutput{
		Value:        amount,
		ScriptPubkey: to,
	})
	if acc > amount {
		outputs = append(outputs, TXOutput{
			Value:        acc - amount,
			ScriptPubkey: from,
		})
	}

	tx := Transaction{
		Vin:  inputs,
		Vout: outputs,
	}
	tx.SetID()

	return &tx
}
