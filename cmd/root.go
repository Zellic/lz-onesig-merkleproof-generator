package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"merkle-cli/merkle"
	"merkle-cli/models"
	"merkle-cli/utils"

	"github.com/spf13/cobra"
)

var (
	oneSigID     uint64
	contractAddr string
	batchFile    string
	verbose      bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "merkle-cli",
	Short: "OneSig Merkle Root Generator",
	Long: `OneSig Merkle Root Generator

A CLI tool for generating Merkle roots for OneSig transaction batches according to the
LayerZero OneSig specification.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if batchFile == "" {
			return fmt.Errorf("transaction batch file is required")
		}

		// Read the transaction batch file
		data, err := os.ReadFile(batchFile)
		if err != nil {
			return fmt.Errorf("failed to read transaction batch file: %w", err)
		}

		// Parse the transaction batch
		var batch models.TransactionBatch
		if err := json.Unmarshal(data, &batch); err != nil {
			return fmt.Errorf("failed to parse transaction batch: %w", err)
		}

		// Handle both new (with groups) and legacy (flat transactions) formats
		var leaves [][]byte
		var nonceToLeaf = make(map[uint64][]byte)
		var nonceToProof = make(map[uint64][][]byte)
		var nonceToCalls = make(map[uint64][]models.Call)

		// Process transaction groups if they exist
		if len(batch.Groups) > 0 {
			for _, group := range batch.Groups {
				// Generate only one leaf for each group's nonce
				if len(group.Calls) > 0 {
					// Check if a leaf already exists for this nonce
					if _, exists := nonceToLeaf[group.Nonce]; !exists {
						// Generate leaf using all calls
						leaf, err := utils.EncodeLeaf(oneSigID, contractAddr, group.Nonce, group.Calls)
						if err != nil {
							return fmt.Errorf("failed to encode leaf for group %d: %w", group.Nonce, err)
						}

						leaves = append(leaves, leaf)
						nonceToLeaf[group.Nonce] = leaf
						nonceToCalls[group.Nonce] = group.Calls
					}
				}
			}
		} else {
			return fmt.Errorf("transaction batch is empty")
		}

		// Ensure we have at least one valid leaf
		if len(leaves) == 0 {
			return fmt.Errorf("no valid transactions found in batch")
		}

		// Sort leaves for consistent merkle root generation
		sortedLeaves := merkle.SortLeaves(leaves)

		// Generate the merkle tree
		tree, err := merkle.NewMerkleTree(sortedLeaves)
		if err != nil {
			return fmt.Errorf("failed to generate merkle tree: %w", err)
		}

		// Output the merkle root
		fmt.Println("Merkle Root:", tree.GetRootHex())

		// Generate proof for each nonce
		for nonce, leaf := range nonceToLeaf {
			proof, err := tree.GenerateProof(leaf)
			if err != nil {
				return fmt.Errorf("failed to generate proof for nonce %d: %w", nonce, err)
			}
			nonceToProof[nonce] = proof
		}

		// Output the proofs if verbose mode is enabled
		if verbose {
			fmt.Println("\nMerkle Proofs by Nonce:")

			// Sort keys to output in nonce order
			var nonces []uint64
			for nonce := range nonceToLeaf {
				nonces = append(nonces, nonce)
			}

			// Sort nonces in ascending order
			for i := 0; i < len(nonces); i++ {
				for j := i + 1; j < len(nonces); j++ {
					if nonces[i] > nonces[j] {
						nonces[i], nonces[j] = nonces[j], nonces[i]
					}
				}
			}

			for _, nonce := range nonces {
				leaf := nonceToLeaf[nonce]
				proof := nonceToProof[nonce]
				calls := nonceToCalls[nonce]

				// Convert proofs to hex for display
				var proofHex []string
				for _, p := range proof {
					proofHex = append(proofHex, fmt.Sprintf("0x%x", p))
				}

				fmt.Printf("\nNonce %d:\n", nonce)
				fmt.Printf("  Calls: %d\n", len(calls))
				fmt.Printf("  Leaf: 0x%x\n", leaf)
				fmt.Printf("  Proof:\n")
				for j, p := range proofHex {
					fmt.Printf("    %d: %s\n", j+1, p)
				}

				// Verify the proof
				isValid := merkle.VerifyProof(tree.Root, leaf, proof)
				fmt.Printf("  Proof Valid: %v\n", isValid)
			}
		}

		return nil
	},
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
	// OneSig ID flag
	rootCmd.Flags().Uint64VarP(&oneSigID, "onesig-id", "o", 0, "OneSig ID (typically chain ID)")
	rootCmd.MarkFlagRequired("onesig-id")

	// Contract address flag
	rootCmd.Flags().StringVarP(&contractAddr, "contract-addr", "c", "", "OneSig contract address (defaults to 0xdEaD if not provided)")

	// Transaction batch file flag
	rootCmd.Flags().StringVarP(&batchFile, "batch-file", "f", "", "Path to transaction batch JSON file")
	rootCmd.MarkFlagRequired("batch-file")

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output including Merkle proofs")
}
