package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/node"
	"github.com/jTanG0506/go-blockchain/wallet"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			miner, _ := cmd.Flags().GetString(flagMiner)
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)

			toshi := database.NewAccount(wallet.ToshiAccount)
			jtang := database.NewAccount(wallet.JTangAccount)
			qudsii := database.NewAccount(wallet.QudsiiAccount)

			peer := node.NewPeerNode(
				"127.0.0.1",
				8080,
				true,
				toshi,
				false,
			)

			n := node.NewNode(getDataDirFromCmd(cmd), ip, port, database.NewAccount(miner), peer)

			n.AddPendingTX(database.NewTx(toshi, toshi, 3, ""), peer)
			n.AddPendingTX(database.NewTx(toshi, jtang, 2000, ""), peer)
			n.AddPendingTX(database.NewTx(jtang, toshi, 1, ""), peer)
			n.AddPendingTX(database.NewTx(jtang, qudsii, 1000, ""), peer)
			n.AddPendingTX(database.NewTx(jtang, toshi, 50, ""), peer)

			ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*15)

			go func() {
				ticker := time.NewTicker(time.Second * 10)
				for {
					select {
					case <-ticker.C:
						if !n.LatestBlockHash().IsEmpty() {
							closeNode()
							return
						}
					}
				}
			}()

			err := n.Run(ctx)
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)
	migrateCmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block rewards")
	migrateCmd.Flags().String(flagIP, node.DefaultIP, "exposed IP for communication with peers")
	migrateCmd.Flags().Uint64(flagPort, node.DefaultHTTPPort, "exposed HTTP port for communication with peers")

	return migrateCmd
}
