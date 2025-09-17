package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rpcURL         string
	blockHeight    int64
	accountAddress string
)

var rootCmd = &cobra.Command{
	Use:   "gethtried",
	Short: "A CLI tool to visualize Ethereum Tries from a Geth archive node.",
	Long: `gethtried is a powerful tool that connects to a Geth archive node 
to fetch and visualize the underlying Merkle Patricia Tries (State, Storage, etc.).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rpcURL, "rpc-url", "http://localhost:8545", "Geth Archive Node RPC URL")
	rootCmd.PersistentFlags().Int64Var(&blockHeight, "block-height", 0, "Block height (required)")
	rootCmd.PersistentFlags().StringVar(&accountAddress, "account-address", "", "Account address to inspect (required)")

	rootCmd.MarkFlagRequired("block-height")
	rootCmd.MarkFlagRequired("account-address")
}
