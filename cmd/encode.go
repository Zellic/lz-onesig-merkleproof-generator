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
	filePath            string
	leafEncodingVersion int
	encodeSortedPairs   bool
	encodeSortLeaves    bool
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode leaves and generate merkle tree from JSON input",
	Long: `Encode leaves and generate merkle tree from JSON input.

This command takes a JSON file containing OneSig leaves, encodes them according to 
the specified version, and generates a merkle tree with proofs for each leaf.

Example:
  merkle-cli encode --file-path input.json --leafEncodingVersion 1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if filePath == "" {
			return fmt.Errorf("file path is required")
		}

		// Read the input file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}

		// Parse the input
		var input models.InputFormat
		if err := json.Unmarshal(data, &input); err != nil {
			return fmt.Errorf("failed to parse input JSON: %w", err)
		}

		// Validate input
		if err := validateInput(input); err != nil {
			return fmt.Errorf("input validation failed: %w", err)
		}

		// Encode leaves
		var encodedLeaves [][]byte
		var leafToOriginal = make(map[string]models.Leaf)

		for _, leaf := range input.Leaves {
			encodedLeaf, err := utils.EncodeLeafV2(leaf, leafEncodingVersion)
			if err != nil {
				return fmt.Errorf("failed to encode leaf (nonce: %s, oneSigId: %s): %w",
					leaf.Nonce, leaf.OneSigId, err)
			}

			encodedLeaves = append(encodedLeaves, encodedLeaf)
			leafToOriginal[fmt.Sprintf("0x%x", encodedLeaf)] = leaf
		}

		// Generate merkle tree with options
		tree, err := merkle.NewMerkleTreeWithOptions(encodedLeaves, merkle.TreeOptions{
			SortedPairs: encodeSortedPairs,
			SortLeaves:  encodeSortLeaves,
		})
		if err != nil {
			return fmt.Errorf("failed to generate merkle tree: %w", err)
		}

		// Generate proofs
		var proofs []models.ProofOutput
		for _, encodedLeaf := range encodedLeaves {
			proof, err := tree.GenerateProof(encodedLeaf)
			if err != nil {
				return fmt.Errorf("failed to generate proof: %w", err)
			}

			// Convert proof to hex strings
			var proofHex []string
			for _, p := range proof {
				proofHex = append(proofHex, fmt.Sprintf("0x%x", p))
			}

			// Get original leaf data
			leafHex := fmt.Sprintf("0x%x", encodedLeaf)
			originalLeaf := leafToOriginal[leafHex]

			proofs = append(proofs, models.ProofOutput{
				Leaf:                leafHex,
				Nonce:               originalLeaf.Nonce,
				OneSigId:            originalLeaf.OneSigId,
				TargetOneSigAddress: originalLeaf.TargetOneSigAddress,
				Proof:               proofHex,
			})
		}

		// Create output
		output := models.OutputFormat{
			MerkleRoot: tree.GetRootHex(),
			Proofs:     proofs,
		}

		// Output as JSON
		outputJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}

		fmt.Println(string(outputJSON))
		return nil
	},
}

// validateInput validates the input according to the requirements
func validateInput(input models.InputFormat) error {
	if len(input.Leaves) == 0 {
		return fmt.Errorf("no leaves provided")
	}

	// Check for duplicate nonces within the same oneSigId
	nonceMap := make(map[string]map[string]bool) // oneSigId -> nonce -> exists

	for _, leaf := range input.Leaves {
		if leaf.OneSigId == "" {
			return fmt.Errorf("oneSigId is required")
		}
		if leaf.Nonce == "" {
			return fmt.Errorf("nonce is required")
		}
		if leaf.TargetOneSigAddress == "" {
			return fmt.Errorf("targetOneSigAddress is required")
		}
		if len(leaf.Calls) == 0 {
			return fmt.Errorf("at least one call is required")
		}

		// Initialize map for this oneSigId if not exists
		if nonceMap[leaf.OneSigId] == nil {
			nonceMap[leaf.OneSigId] = make(map[string]bool)
		}

		// Check for duplicate nonce within the same oneSigId
		if nonceMap[leaf.OneSigId][leaf.Nonce] {
			return fmt.Errorf("duplicate nonce %s found for oneSigId %s", leaf.Nonce, leaf.OneSigId)
		}

		nonceMap[leaf.OneSigId][leaf.Nonce] = true

		// Validate calls
		for i, call := range leaf.Calls {
			if call.To == "" {
				return fmt.Errorf("call %d: 'to' address is required", i)
			}
			if call.Value == nil {
				return fmt.Errorf("call %d: 'value' is required", i)
			}
		}
	}

	return nil
}

func init() {
	encodeCmd.Flags().StringVarP(&filePath, "file-path", "f", "", "Path to the JSON file containing the leaves")
	encodeCmd.MarkFlagRequired("file-path")

	encodeCmd.Flags().IntVarP(&leafEncodingVersion, "leafEncodingVersion", "v", 1, "Specifies the encoding version to use for the leaves")
	encodeCmd.Flags().BoolVar(&encodeSortedPairs, "sortedPairs", true, "Use sorted pairs when building the Merkle Tree (default: false, matching MerkleTreeJs)")
	encodeCmd.Flags().BoolVar(&encodeSortLeaves, "sortLeaves", false, "Sort leaves before building the Merkle Tree (default: false, matching MerkleTreeJs)")
}
