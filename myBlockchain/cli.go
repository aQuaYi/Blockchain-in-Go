package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// CLI 负责处理从命令行接收的指令
type CLI struct {
	bc *blockchain
}

// Run 处理命令行的指令
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block Data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) validateArgs() {

	return
}

func (cli *CLI) printUsage() {

	return
}

func (cli *CLI) addBlock(data string) {
	cli.bc.addBlock(data)
	fmt.Println("Success added!")
}

func (cli *CLI) printChain() {
	bi := cli.bc.Iterator()

	// TODO: 删除此处输出}
	fmt.Println(bi.currentHash)

	for len(bi.currentHash) != 0 {
		b := bi.Next()

		// TODO: 删除此处输出}
		fmt.Println(b)

		fmt.Printf(" Prev. hash	: %s\n", b.preBlockHash)
		fmt.Printf("		data:%s\n", b.data)
		pow := newProofOfWork(b)
		fmt.Printf("		Pow	: %t\n", pow.validate())
		fmt.Printf("  this hash	:%s\n", b.hash)
		fmt.Println("")
	}
}
