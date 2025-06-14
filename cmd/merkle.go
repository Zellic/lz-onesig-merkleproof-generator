package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"merkle-cli/merkle"

	"github.com/spf13/cobra"
)

var (
	encodedInput string
	sortedPairs  bool
	sortLeaves   bool
)

// merkleCmd represents the merkle command
var merkleCmd = &cobra.Command{
	Use:   "merkle",
	Short: "Generate merkle tree from pre-encoded leaves",
	Long: `Generate merkle tree from pre-encoded leaves.

This command takes either a file path to a JSON array of encoded leaves or 
a comma-separated array of encoded leaves and generates a merkle tree.

Examples:
  # From file
  merkle-cli merkle --encodedInput encoded.json --sortedPairs true

  # From comma-separated values
  merkle-cli merkle --encodedInput "0xabc...,0xdef..." --sortedPairs false`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if encodedInput == "" {
			return fmt.Errorf("encoded input is required")
		}

		// Create merkle module
		module := merkle.NewMerkleModule()
		options := merkle.MerkleOptions{
			SortedPairs: sortedPairs,
			SortLeaves:  sortLeaves,
		}

		var result *merkle.MerkleResult
		var err error

		// Determine input type and process accordingly
		if strings.Contains(encodedInput, ",") {
			// Comma-separated values
			result, err = module.GenerateFromEncodedString(encodedInput, options)
		} else {
			// Try to read as file first
			if _, statErr := os.Stat(encodedInput); statErr == nil {
				// File exists, read it
				result, err = module.GenerateFromEncodedFile(encodedInput, options)
			} else {
				// Treat as single encoded leaf
				result, err = module.GenerateFromEncodedString(encodedInput, options)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to generate merkle tree: %w", err)
		}

		// Output as JSON
		outputJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}

		fmt.Println(string(outputJSON))
		return nil
	},
}

func init() {
	merkleCmd.Flags().StringVarP(&encodedInput, "encodedInput", "i", "", "File path to JSON array of encoded leaves or comma-separated array")
	merkleCmd.MarkFlagRequired("encodedInput")

	merkleCmd.Flags().BoolVarP(&sortedPairs, "sortedPairs", "s", false, "Use sorted pairs when building the Merkle Tree (default: false, matching MerkleTreeJs)")
	merkleCmd.Flags().BoolVarP(&sortLeaves, "sortLeaves", "l", false, "Sort leaves before building the Merkle Tree (default: false, matching MerkleTreeJs)")
}
