package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
	// 使用 address 创建区块链
	// 创世区块的 coinbase 交易的奖励，发送给 address
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("Done!")
}

// 获取 address 中的账户余额
func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0
	// 获取 address 所有的未花费的输出
	UTXOs := bc.FindUTXO(address)

	// 统计所有未花费的输出到 balance
	for _, out := range UTXOs {
		balance += out.Value
	}

	// 在命令行输出结果
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// 在命令行输出本程序的用法
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// 验证本程序附加参数的正确性
// 个数不能少于 2
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// 打印输出区块链
func (cli *CLI) printChain() {
	// TODO: Fix this  // fix what??? by aQua
	// 新建一个区块链实例
	bc := NewBlockchain("")
	defer bc.db.Close()

	// 生成区块链的迭代器
	bci := bc.Iterator()

	for {
		// 从迭代器获取下一个区块
		block := bci.Next()

		// 输出各项内容
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("      Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("       PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		// 下一个区块不存在的话，结束 for 循环
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

// 区块链的发送虚拟币功能
// from 把 amount 个虚拟币，发送给 to
func (cli *CLI) send(from, to string, amount int) {
	// 创建区块链
	bc := NewBlockchain(from)
	defer bc.db.Close()

	// 创建新的发送交易
	tx := NewUTXOTransaction(from, to, amount, bc)
	// 使用交易进行挖矿
	bc.MineBlock([]*Transaction{tx})
	// 在命令行反馈成功发送
	fmt.Println("Success!")
}

// Run parses command line arguments and processes commands
// 启动 cli
func (cli *CLI) Run() {
	// 验证 cli 运行条件
	cli.validateArgs()

	// 获取账户余额命令
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	// 创建区块链命令
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	// 发送虚拟币命令
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	// 打印输出区块链命令
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	// 命令的参数
	//
	// getBalance 命令获取的 address 变量放在 getBalanceAddress 变量中
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	// createBlockchain 命令获取的 address 变量放在 CreateBlockchainAdress 变量中
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	// send 命令的 from 变量放在 sendFrom 中
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	// send 命令的 to 变量放在 sendTo 中
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	// send 命令的 amount 变量放在 sendAmount 中
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// 分别解析各个命令
	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		// 如果 getBalanceAddress 是 "" 的话
		if *getBalanceAddress == "" {
			// 输出 getBalance 的用法
			getBalanceCmd.Usage()
			// 再退出程序
			os.Exit(1)
		}
		// 执行 getBalance 命令
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		// 如果 createBlockchainAddress 是 "" 的话
		if *createBlockchainAddress == "" {
			// 输出 CreateBlockchain 的用法
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		// 执行 createBlockchain 命令
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		// 如果没有全部满足 send 命令的要求
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			// 输出 send 命令的用法
			sendCmd.Usage()
			os.Exit(1)
		}

		// 执行 send 命令
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
