package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/node"
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

			block0 := node.NewPendingBlock(
				database.Hash{},
				state.NextBlockNumber(),
				[]database.Tx{
					database.NewTx("toshi", "toshi", 3, ""),
					database.NewTx("toshi", "toshi", 700, "reward"),
					database.NewTx("toshi", "jtang", 2000, ""),
					database.NewTx("toshi", "toshi", 100, "reward"),
					database.NewTx("jtang", "toshi", 1, ""),
					database.NewTx("jtang", "qudsii", 1000, ""),
					database.NewTx("jtang", "toshi", 50, ""),
					database.NewTx("toshi", "toshi", 600, "reward"),
					database.NewTx("toshi", "toshi", 24700, "reward"),
				},
			)

			_, err = node.Mine(context.Background(), block0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
