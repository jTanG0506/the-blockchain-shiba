package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jTanG0506/go-blockchain/database"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			block0 := database.NewBlock(
				database.Hash{},
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("toshi", "toshi", 3, ""),
					database.NewTx("toshi", "toshi", 700, "reward"),
				},
			)

			block0hash, err := state.AddBlock(block0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block1 := database.NewBlock(
				block0hash,
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("toshi", "jtang", 2000, ""),
					database.NewTx("toshi", "toshi", 100, "reward"),
					database.NewTx("jtang", "toshi", 1, ""),
					database.NewTx("jtang", "qudsii", 1000, ""),
					database.NewTx("jtang", "toshi", 50, ""),
					database.NewTx("toshi", "toshi", 600, "reward"),
				},
			)

			block1hash, err := state.AddBlock(block1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block2 := database.NewBlock(
				block1hash,
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("toshi", "toshi", 24700, "reward"),
				},
			)

			_, err = state.AddBlock(block2)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
