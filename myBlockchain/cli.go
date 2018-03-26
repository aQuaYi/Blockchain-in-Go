package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) addBlock(data string) {
	cli.bc.addBlock(data)
	fmt.Println("Success added!")
}

func (cli *CLI) printChain() {
	fmt.Println("Begin printchain...")

	bi := cli.bc.Iterator()
	for {
		block := bi.Next()

		fmt.Printf("Prev. hash: %x\n", block.preBlockHash)
		fmt.Printf("Data: %s\n", block.data)
		fmt.Printf("Hash: %x\n", block.hash)
		pow := newProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.validate()))
		fmt.Println()

		if len(block.preBlockHash) == 0 {
			break
		}
	}
}
