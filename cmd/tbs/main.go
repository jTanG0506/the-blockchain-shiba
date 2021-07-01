package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var tbsCmd = &cobra.Command{
		Use:   "tbs",
		Short: "The Blockchain Shiba CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	tbsCmd.AddCommand(versionCmd)
	tbsCmd.AddCommand(balancesCmd())
	tbsCmd.AddCommand(txCmd())

	err := tbsCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
