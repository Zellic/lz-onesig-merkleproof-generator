package merkle

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"merkle-cli/models"
	"merkle-cli/utils"
)

// MerkleModule provides high-level merkle tree operations
type MerkleModule struct{}

// MerkleOptions represents options for merkle tree generation
type MerkleOptions struct {
	SortedPairs bool `json:"sortedPairs"`
	SortLeaves  bool `json:"sortLeaves"`
}

// MerkleResult represents the result of merkle tree generation
type MerkleResult struct {
	MerkleRoot string               `json:"merkleRoot"`
	Proofs     []models.ProofOutput `json:"proofs"`
}

// NewMerkleModule creates a new merkle module instance
func NewMerkleModule() *MerkleModule {
	return &MerkleModule{}
}

// GenerateFromEncodedLeaves generates merkle tree from pre-encoded leaves
func (m *MerkleModule) GenerateFromEncodedLeaves(encodedLeaves []string, options MerkleOptions) (*MerkleResult, error) {
	// Convert hex strings to bytes
	var leaves [][]byte
	for i, hexLeaf := range encodedLeaves {
		leafBytes, err := utils.HexToBytes(hexLeaf)
		if err != nil {
			return nil, fmt.Errorf("invalid hex string at index %d: %w", i, err)
		}
		leaves = append(leaves, leafBytes)
	}

	return m.GenerateFromLeaves(leaves, options)
}

// GenerateFromEncodedFile generates merkle tree from a file containing encoded leaves
func (m *MerkleModule) GenerateFromEncodedFile(filePath string, options MerkleOptions) (*MerkleResult, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var input models.EncodedLeavesInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return m.GenerateFromEncodedLeaves(input.EncodedLeaves, options)
}

// GenerateFromEncodedString generates merkle tree from comma-separated encoded leaves
func (m *MerkleModule) GenerateFromEncodedString(encodedString string, options MerkleOptions) (*MerkleResult, error) {
	// Split by comma and trim spaces
	encodedLeaves := strings.Split(encodedString, ",")
	for i, leaf := range encodedLeaves {
		encodedLeaves[i] = strings.TrimSpace(leaf)
	}

	return m.GenerateFromEncodedLeaves(encodedLeaves, options)
}

// GenerateFromLeaves generates merkle tree from raw leaf bytes
func (m *MerkleModule) GenerateFromLeaves(leaves [][]byte, options MerkleOptions) (*MerkleResult, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("no leaves provided")
	}

	// Create merkle tree with options
	tree, err := NewMerkleTreeWithOptions(leaves, TreeOptions{
		SortedPairs: options.SortedPairs,
		SortLeaves:  options.SortLeaves,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create merkle tree: %w", err)
	}

	// Generate proofs for all leaves
	var proofs []models.ProofOutput
	for _, leaf := range leaves {
		proof, err := tree.GenerateProof(leaf)
		if err != nil {
			return nil, fmt.Errorf("failed to generate proof: %w", err)
		}

		// Convert proof to hex strings
		var proofHex []string
		for _, p := range proof {
			proofHex = append(proofHex, fmt.Sprintf("0x%x", p))
		}

		proofs = append(proofs, models.ProofOutput{
			Leaf:                fmt.Sprintf("0x%x", leaf),
			Nonce:               "", // Not available in merkle-only mode
			OneSigId:            "", // Not available in merkle-only mode
			TargetOneSigAddress: "", // Not available in merkle-only mode
			Proof:               proofHex,
		})
	}

	return &MerkleResult{
		MerkleRoot: tree.GetRootHex(),
		Proofs:     proofs,
	}, nil
}

// VerifyProof verifies a merkle proof
func (m *MerkleModule) VerifyProof(root string, leaf string, proof []string, options MerkleOptions) (bool, error) {
	// Convert hex strings to bytes
	rootBytes, err := utils.HexToBytes(root)
	if err != nil {
		return false, fmt.Errorf("invalid root hex: %w", err)
	}

	leafBytes, err := utils.HexToBytes(leaf)
	if err != nil {
		return false, fmt.Errorf("invalid leaf hex: %w", err)
	}

	var proofBytes [][]byte
	for _, p := range proof {
		pBytes, err := utils.HexToBytes(p)
		if err != nil {
			return false, fmt.Errorf("invalid proof element hex: %w", err)
		}
		proofBytes = append(proofBytes, pBytes)
	}

	// Verify using tree options
	return VerifyProofWithOptions(rootBytes, leafBytes, proofBytes, TreeOptions{
		SortedPairs: options.SortedPairs,
		SortLeaves:  options.SortLeaves,
	}), nil
}
