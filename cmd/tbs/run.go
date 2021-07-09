package main

import (
	"fmt"
	"os"

	"github.com/jTanG0506/go-blockchain/node"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the TBS node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetUint64(flagPort)

			fmt.Println("Launching TBS node and its HTTP API...")

			bootstrap := node.NewPeerNode(
				"BOOTSTRAP_NODE_IP",
				8080,
				true,
				true,
			)

			n := node.NewNode(getDataDirFromCmd(cmd), port, bootstrap)
			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().Uint64(flagPort, node.DefaultHTTPPort, "exposed HTTP port for communication with peers")

	return runCmd
}
