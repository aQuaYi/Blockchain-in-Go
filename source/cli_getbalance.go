package main

import (
	"fmt"
	"log"
)

// 获取指定地址的账户余额
func (cli *CLI) getBalance(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for i := range UTXOs {
		balance += UTXOs[i].Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
