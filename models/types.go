package models

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Call represents a single call to be executed
type Call struct {
	To    string  `json:"to"`
	Value *BigInt `json:"value"`
	Data  string  `json:"data"`
}

// BigInt wraps *big.Int to provide custom JSON marshaling/unmarshaling
type BigInt struct {
	*big.Int
}

// NewBigInt creates a new BigInt from a *big.Int
func NewBigInt(i *big.Int) *BigInt {
	if i == nil {
		return &BigInt{big.NewInt(0)}
	}
	return &BigInt{i}
}

// UnmarshalJSON implements json.Unmarshaler for BigInt
func (b *BigInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try unmarshaling as number
		var n json.Number
		if err := json.Unmarshal(data, &n); err != nil {
			return fmt.Errorf("cannot unmarshal %s into BigInt", string(data))
		}
		s = string(n)
	}

	// Parse the string value
	value, err := parseBigIntString(s)
	if err != nil {
		return err
	}

	b.Int = value
	return nil
}

// MarshalJSON implements json.Marshaler for BigInt
func (b *BigInt) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return json.Marshal("0")
	}
	return json.Marshal(b.Int.String())
}

// parseBigIntString parses a string that can be either a decimal number or hex string into *big.Int
func parseBigIntString(s string) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}

	// Try parsing as hex first
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		result := new(big.Int)
		result, ok := result.SetString(s[2:], 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex number: %s", s)
		}
		return result, nil
	}

	// Try parsing as decimal
	if val, err := strconv.ParseUint(s, 10, 64); err == nil {
		return big.NewInt(int64(val)), nil
	}

	// Try parsing as big decimal
	result := new(big.Int)
	result, ok := result.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid number: %s", s)
	}

	return result, nil
}

// Leaf represents a single leaf in the merkle tree input
type Leaf struct {
	Nonce               string `json:"nonce"`    // bigint as string/number
	OneSigId            string `json:"oneSigId"` // bigint as string/number
	TargetOneSigAddress string `json:"targetOneSigAddress"`
	Calls               []Call `json:"calls"`
}

// InputFormat represents the JSON input format for leaf encoding
type InputFormat struct {
	Leaves []Leaf `json:"leaves"`
}

// ProofOutput represents a single proof in the output
type ProofOutput struct {
	Leaf                string   `json:"leaf"`     // hashed encoded leaf
	Nonce               string   `json:"nonce"`    // bigint as string
	OneSigId            string   `json:"oneSigId"` // bigint as string
	TargetOneSigAddress string   `json:"targetOneSigAddress"`
	Proof               []string `json:"proof"` // array of hex strings
}

// OutputFormat represents the JSON output format
type OutputFormat struct {
	MerkleRoot string        `json:"merkleRoot"`
	Proofs     []ProofOutput `json:"proofs"`
}

// EncodedLeavesInput represents input for merkle-only mode
type EncodedLeavesInput struct {
	EncodedLeaves []string `json:"encodedLeaves"` // array of hex-encoded leaves
}

// Legacy types (keeping for backward compatibility during transition)
// TODO: Remove these after full migration

// Transaction represents a batch of calls to be executed atomically
type Transaction struct {
	Nonce uint64 `json:"nonce"`
	Calls []Call `json:"calls"`
}

// TransactionGroup represents a group of calls that share the same nonce
type TransactionGroup struct {
	Nonce uint64 `json:"nonce"`
	Calls []Call `json:"calls"`
}

// TransactionBatch represents a collection of transaction groups to be merklized
type TransactionBatch struct {
	Groups []TransactionGroup `json:"groups"`
}
