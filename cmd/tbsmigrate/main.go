package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jTanG0506/go-blockchain/database"
)

func main() {
	state, err := database.NewStateFromDisk()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("toshi", "jtang", 100, ""),
			database.NewTx("toshi", "toshi", 50, "reward"),
		},
	)

	state.AddBlock(block0)
	block0Hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0Hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("toshi", "jtang", 100, ""),
			database.NewTx("toshi", "toshi", 50, "reward"),
		},
	)

	state.AddBlock(block1)
	state.Persist()
}
