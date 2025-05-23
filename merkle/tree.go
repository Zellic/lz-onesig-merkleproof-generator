package merkle

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/crypto"
)

// MerkleTree implements a binary Merkle tree
type MerkleTree struct {
	Root  []byte
	Leafs [][]byte
}

// NewMerkleTree creates a new Merkle tree from a set of leaves
func NewMerkleTree(leaves [][]byte) (*MerkleTree, error) {
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

	// Build the Merkle tree
	root, err := buildTree(leafCopies)
	if err != nil {
		return nil, err
	}

	return &MerkleTree{
		Root:  root,
		Leafs: leafCopies,
	}, nil
}

// buildTree builds the Merkle tree from the leaves and returns the root hash
func buildTree(leaves [][]byte) ([]byte, error) {
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
		// If we have an odd number of leaves, duplicate the last one
		if i+1 == len(leaves) {
			nextLevel = append(nextLevel, hashPair(leaves[i], leaves[i]))
		} else {
			nextLevel = append(nextLevel, hashPair(leaves[i], leaves[i+1]))
		}
	}

	// Recursively build the next level
	return buildTree(nextLevel)
}

// hashPair hashes two leaves together to form a parent node
func hashPair(left, right []byte) []byte {
	// Sort the pair to ensure consistent hashing order
	if bytes.Compare(left, right) > 0 {
		left, right = right, left
	}

	// Concatenate and hash
	concat := append(left, right...)
	return crypto.Keccak256(concat)
}

// VerifyProof verifies a Merkle proof for a specific leaf
func VerifyProof(root []byte, leaf []byte, proof [][]byte) bool {
	currentHash := leaf

	for _, proofElement := range proof {
		currentHash = hashPair(currentHash, proofElement)
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

	return generateProofHelper(m.Leafs, leafIndex), nil
}

// generateProofHelper recursively builds the proof for a leaf at a given index
func generateProofHelper(nodes [][]byte, index int) [][]byte {
	if len(nodes) == 1 {
		return [][]byte{}
	}

	var proof [][]byte
	var nextLevel [][]byte

	// Process pairs of nodes
	for i := 0; i < len(nodes); i += 2 {
		if i+1 == len(nodes) {
			// If we have an odd number of nodes, duplicate the last one
			nextLevel = append(nextLevel, hashPair(nodes[i], nodes[i]))

			if i == index || i+1 == index {
				proof = append(proof, nodes[i])
			}
		} else {
			nextLevel = append(nextLevel, hashPair(nodes[i], nodes[i+1]))

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
	return append(proof, generateProofHelper(nextLevel, nextIndex)...)
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
