package merkle

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/crypto"
)

// TreeOptions represents options for merkle tree construction
type TreeOptions struct {
	SortedPairs bool // Whether to use sorted pairs when hashing (default: false)
	SortLeaves  bool // Whether to sort leaves before building tree (default: false)
}

// MerkleTree implements a binary Merkle tree
type MerkleTree struct {
	Root    []byte
	Leafs   [][]byte
	Options TreeOptions
}

// NewMerkleTree creates a new Merkle tree from a set of leaves with default options (matching MerkleTreeJs defaults)
func NewMerkleTree(leaves [][]byte) (*MerkleTree, error) {
	return NewMerkleTreeWithOptions(leaves, TreeOptions{
		SortedPairs: false,
		SortLeaves:  false,
	})
}

// NewMerkleTreeWithOptions creates a new Merkle tree from a set of leaves with specified options
func NewMerkleTreeWithOptions(leaves [][]byte, options TreeOptions) (*MerkleTree, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("cannot create Merkle tree with no leaves")
	}

	// Create a copy of leaves to avoid modifying the input
	leafCopies := make([][]byte, len(leaves))
	for i, leaf := range leaves {
		leafCopy := make([]byte, len(leaf))
		copy(leafCopy, leaf)
		leafCopies[i] = leafCopy
	}

	// Sort leaves if sortLeaves option is enabled
	if options.SortLeaves {
		sort.Slice(leafCopies, func(i, j int) bool {
			return bytes.Compare(leafCopies[i], leafCopies[j]) < 0
		})
	}

	// Build the Merkle tree
	root, err := buildTreeWithOptions(leafCopies, options)
	if err != nil {
		return nil, err
	}

	return &MerkleTree{
		Root:    root,
		Leafs:   leafCopies,
		Options: options,
	}, nil
}

// buildTreeWithOptions builds the Merkle tree from the leaves and returns the root hash
func buildTreeWithOptions(leaves [][]byte, options TreeOptions) ([]byte, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("cannot build tree with no leaves")
	}

	// If there's only one leaf, it's the root
	if len(leaves) == 1 {
		return leaves[0], nil
	}

	// Create a new level of nodes
	var nextLevel [][]byte

	// Process pairs of leaves
	for i := 0; i < len(leaves); i += 2 {
		// If we have an odd number of leaves, promote the last one to next level
		if i+1 == len(leaves) {
			nextLevel = append(nextLevel, leaves[i])
		} else {
			nextLevel = append(nextLevel, hashPairWithOptions(leaves[i], leaves[i+1], options))
		}
	}

	// Recursively build the next level
	return buildTreeWithOptions(nextLevel, options)
}

// buildTree builds the Merkle tree from the leaves and returns the root hash (legacy function)
func buildTree(leaves [][]byte) ([]byte, error) {
	return buildTreeWithOptions(leaves, TreeOptions{
		SortedPairs: false,
		SortLeaves:  false,
	})
}

// hashPairWithOptions hashes two leaves together to form a parent node with options
func hashPairWithOptions(left, right []byte, options TreeOptions) []byte {
	// Sort the pair if sortedPairs option is enabled
	if options.SortedPairs && bytes.Compare(left, right) > 0 {
		left, right = right, left
	}

	// Concatenate and hash
	concat := append(left, right...)
	return crypto.Keccak256(concat)
}

// hashPair hashes two leaves together to form a parent node (legacy function)
func hashPair(left, right []byte) []byte {
	return hashPairWithOptions(left, right, TreeOptions{
		SortedPairs: false,
		SortLeaves:  false,
	})
}

// VerifyProof verifies a Merkle proof for a specific leaf
func VerifyProof(root []byte, leaf []byte, proof [][]byte) bool {
	currentHash := leaf

	for _, proofElement := range proof {
		currentHash = hashPair(currentHash, proofElement)
	}

	return bytes.Equal(currentHash, root)
}

// VerifyProofWithOptions verifies a Merkle proof for a specific leaf with options
func VerifyProofWithOptions(root []byte, leaf []byte, proof [][]byte, options TreeOptions) bool {
	currentHash := leaf

	for _, proofElement := range proof {
		currentHash = hashPairWithOptions(currentHash, proofElement, options)
	}

	return bytes.Equal(currentHash, root)
}

// GenerateProof generates a Merkle proof for a specific leaf
func (m *MerkleTree) GenerateProof(leaf []byte) ([][]byte, error) {
	// Find the leaf index
	leafIndex := -1
	for i, l := range m.Leafs {
		if bytes.Equal(l, leaf) {
			leafIndex = i
			break
		}
	}

	if leafIndex == -1 {
		return nil, fmt.Errorf("leaf not found in tree")
	}

	return generateProofHelperWithOptions(m.Leafs, leafIndex, m.Options), nil
}

// generateProofHelperWithOptions recursively builds the proof for a leaf at a given index with options
func generateProofHelperWithOptions(nodes [][]byte, index int, options TreeOptions) [][]byte {
	if len(nodes) == 1 {
		return [][]byte{}
	}

	var proof [][]byte
	var nextLevel [][]byte

	// Process pairs of nodes
	for i := 0; i < len(nodes); i += 2 {
		if i+1 == len(nodes) {
			// If we have an odd number of nodes, promote the last one to next level
			nextLevel = append(nextLevel, nodes[i])

			// No proof element needed for promoted node
		} else {
			nextLevel = append(nextLevel, hashPairWithOptions(nodes[i], nodes[i+1], options))

			if i == index {
				proof = append(proof, nodes[i+1])
			} else if i+1 == index {
				proof = append(proof, nodes[i])
			}
		}
	}

	// Calculate the index for the next level
	nextIndex := index / 2

	// Recursively build the rest of the proof
	return append(proof, generateProofHelperWithOptions(nextLevel, nextIndex, options)...)
}

// generateProofHelper recursively builds the proof for a leaf at a given index (legacy function)
func generateProofHelper(nodes [][]byte, index int) [][]byte {
	return generateProofHelperWithOptions(nodes, index, TreeOptions{
		SortedPairs: false,
		SortLeaves:  false,
	})
}

// GetRootHex returns the root hash as a hexadecimal string
func (m *MerkleTree) GetRootHex() string {
	return "0x" + hex.EncodeToString(m.Root)
}

// SortLeaves sorts the leaves for consistent tree generation
func SortLeaves(leaves [][]byte) [][]byte {
	sortedLeaves := make([][]byte, len(leaves))
	copy(sortedLeaves, leaves)

	sort.Slice(sortedLeaves, func(i, j int) bool {
		return bytes.Compare(sortedLeaves[i], sortedLeaves[j]) < 0
	})

	return sortedLeaves
}
