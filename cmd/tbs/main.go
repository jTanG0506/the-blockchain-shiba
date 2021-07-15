package main

import (
	"fmt"
	"os"

	"github.com/jTanG0506/go-blockchain/fs"
	"github.com/spf13/cobra"
)

const flagKeystoreFile = "keystore"
const flagDataDir = "datadir"
const flagMiner = "miner"
const flagIP = "ip"
const flagPort = "port"
const flagBootstrapAcc = "bootstrap-account"
const flagBootstrapIp = "bootstrap-ip"
const flagBootstrapPort = "bootstrap-port"

func main() {
	var tbsCmd = &cobra.Command{
		Use:   "tbs",
		Short: "The Blockchain Shiba CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	tbsCmd.AddCommand(versionCmd)
	tbsCmd.AddCommand(migrateCmd())
	tbsCmd.AddCommand(walletCmd())
	tbsCmd.AddCommand(runCmd())
	tbsCmd.AddCommand(balancesCmd())

	err := tbsCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data dit where the DB will be stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func addKeystoreFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagKeystoreFile, "", "Absolute path to the encrypted keystore file")
	cmd.MarkFlagRequired(flagKeystoreFile)
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)
	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
