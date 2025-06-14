package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "merkle-cli",
	Short: "OneSig Merkle Tree Generator",
	Long: `OneSig Merkle Tree Generator

A CLI tool for generating Merkle trees and proofs for OneSig transaction batches 
according to the LayerZero OneSig specification.

The tool supports two main operations:
1. encode: Process JSON input to encode leaves and generate merkle tree with proofs
2. merkle: Generate merkle tree from pre-encoded leaves

Examples:
  # Generate merkle tree from JSON input
  merkle-cli encode --file-path input.json --leafEncodingVersion 1

  # Generate merkle tree from encoded leaves
  merkle-cli merkle --encodedInput encoded.json`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(encodeCmd)
	rootCmd.AddCommand(merkleCmd)
}
