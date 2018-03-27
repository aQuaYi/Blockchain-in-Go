package main

import (
	"fmt"
)

const subsidy = 10

// Transaction 是交易
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
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
