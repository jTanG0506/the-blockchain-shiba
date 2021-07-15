package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/jTanG0506/go-blockchain/wallet"
	"github.com/spf13/cobra"
)

func walletCmd() *cobra.Command {
	var walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "Manages blockchain accounts and keys.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	walletCmd.AddCommand(walletNewAccountCmd())
	walletCmd.AddCommand(walletPrintPrivateKeyCmd())
	return walletCmd
}

func walletNewAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "new-account",
		Short: "Creates a new account with a set of elliptic-curve private and public keys",
		Run: func(cmd *cobra.Command, args []string) {
			password := getPassphrase("Enter a password to encrypt the new wallet:", true)
			dataDir := getDataDirFromCmd(cmd)

			ks := keystore.NewKeyStore(
				wallet.GetKeystoreDirPath(dataDir),
				keystore.StandardScryptN,
				keystore.StandardScryptP,
			)
			acc, err := ks.NewAccount(password)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("New account created: %s\n", acc.Address.Hex())
		},
	}

	addDefaultRequiredFlags(cmd)
	return cmd
}

func walletPrintPrivateKeyCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "pk-print",
		Short: "Unlocks keystore file and prints the private and public keypair",
		Run: func(cmd *cobra.Command, args []string) {
			ksFile, _ := cmd.Flags().GetString(flagKeystoreFile)
			password := getPassphrase("Please enter a password to decrypt the wallet:", false)

			keyJson, err := ioutil.ReadFile(ksFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			key, err := keystore.DecryptKey(keyJson, password)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			spew.Dump(key)
		},
	}

	addKeystoreFlag(cmd)
	return cmd
}

func getPassphrase(prompt string, confirm bool) string {
	return utils.GetPassPhrase(prompt, confirm)
}
