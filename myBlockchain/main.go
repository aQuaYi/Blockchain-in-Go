package main

func main() {
	// bc := NewBlockchain()
	var bc *Blockchain
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
