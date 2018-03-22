package main

func main() {
	bc := newBlockchain()

	bc.addBlock("Send 1 BTC to Alice")
	bc.addBlock("Send 2 BTC to Bob")
	bc.addBlock("Send 3 BTC to Candy")

	bc.print()
}
