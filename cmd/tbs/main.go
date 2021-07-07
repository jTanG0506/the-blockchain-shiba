package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"

func main() {
	var tbsCmd = &cobra.Command{
		Use:   "tbs",
		Short: "The Blockchain Shiba CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	tbsCmd.AddCommand(versionCmd)
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

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
