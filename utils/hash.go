package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"merkle-cli/models"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// LeafEncodingVersion is the version byte for the leaf encoding
	LeafEncodingVersion byte = 1
)

// EncodeLeafV2 encodes a leaf according to the new data model and specified version
func EncodeLeafV2(leaf models.Leaf, version int) ([]byte, error) {
	switch version {
	case 1:
		return encodeLeafV1(leaf)
	default:
		return nil, fmt.Errorf("unsupported leaf encoding version: %d", version)
	}
}

// encodeLeafV1 implements version 1 of leaf encoding
func encodeLeafV1(leaf models.Leaf) ([]byte, error) {
	// Parse oneSigId
	oneSigID, err := parseBigInt(leaf.OneSigId)
	if err != nil {
		return nil, fmt.Errorf("invalid oneSigId: %w", err)
	}

	// Parse nonce
	nonce, err := parseBigInt(leaf.Nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce: %w", err)
	}

	// Convert targetOneSigAddress
	addr := common.HexToAddress(leaf.TargetOneSigAddress)
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)

	// Encode oneSigID as 8 bytes (assuming it fits in uint64)
	oneSigIDBytes := make([]byte, 8)
	oneSigIDUint64 := oneSigID.Uint64()
	for i := 0; i < 8; i++ {
		oneSigIDBytes[7-i] = byte(oneSigIDUint64 >> (i * 8))
	}

	// Encode nonce as 8 bytes (assuming it fits in uint64)
	nonceBytes := make([]byte, 8)
	nonceUint64 := nonce.Uint64()
	for i := 0; i < 8; i++ {
		nonceBytes[7-i] = byte(nonceUint64 >> (i * 8))
	}

	// Create ABI definition identical to Solidity's Call struct
	callsAbi, err := abi.JSON(strings.NewReader(`[
		{
			"name": "encodeCalls",
			"type": "function",
			"inputs": [
				{
					"name": "calls",
					"type": "tuple[]",
					"components": [
						{
							"name": "to",
							"type": "address"
						},
						{
							"name": "value",
							"type": "uint256"
						},
						{
							"name": "data",
							"type": "bytes"
						}
					]
				}
			]
		}
	]`))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Convert Go struct to Solidity struct format
	var callsForAbi []struct {
		To    common.Address
		Value *big.Int
		Data  []byte
	}

	for _, call := range leaf.Calls {
		callData, err := HexToBytes(call.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid call data: %w", err)
		}

		// Handle BigInt value
		var value *big.Int
		if call.Value != nil && call.Value.Int != nil {
			value = call.Value.Int
		} else {
			value = big.NewInt(0)
		}

		callsForAbi = append(callsForAbi, struct {
			To    common.Address
			Value *big.Int
			Data  []byte
		}{
			To:    common.HexToAddress(call.To),
			Value: value,
			Data:  callData,
		})
	}

	// Perform ABI encoding (equivalent to abi.encode(_calls))
	callsEncoded, err := callsAbi.Methods["encodeCalls"].Inputs.Pack(callsForAbi)
	if err != nil {
		return nil, fmt.Errorf("failed to encode calls: %w", err)
	}

	// Implementation of abi.encodePacked
	// Equivalent to Solidity's abi.encodePacked(LEAF_ENCODING_VERSION, ONE_SIG_ID, address(this), _nonce, abi.encode(_calls))
	leafData := []byte{LeafEncodingVersion}
	leafData = append(leafData, oneSigIDBytes...) // 8 bytes
	leafData = append(leafData, addrBytes...)     // 32 bytes
	leafData = append(leafData, nonceBytes...)    // 8 bytes
	leafData = append(leafData, callsEncoded...)  // abi.encode(_calls)

	// Double hash leaf data (equivalent to Solidity's keccak256(keccak256(...)))
	firstHash := crypto.Keccak256(leafData)
	finalHash := crypto.Keccak256(firstHash)

	return finalHash, nil
}

// parseBigInt parses a string that can be either a decimal number or hex string into *big.Int
func parseBigInt(s string) (*big.Int, error) {
	if s == "" {
		return nil, fmt.Errorf("empty string")
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

// EncodeLeaf encodes a transaction as a leaf according to OneSig spec (legacy function)
func EncodeLeaf(oneSigID uint64, contractAddr string, nonce uint64, calls []models.Call) ([]byte, error) {
	// Convert contract address
	var addr common.Address
	if contractAddr == "" {
		// Use fixed contract address 0xdEaD as default
		addr = common.HexToAddress("0xdEaD")
	} else {
		// Use user-specified contract address
		addr = common.HexToAddress(contractAddr)
	}

	// Convert address to bytes32 (pad to 32 bytes)
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)

	// Encode oneSigID as 8 bytes
	oneSigIDBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		oneSigIDBytes[7-i] = byte(oneSigID >> (i * 8))
	}

	// Encode nonce as 8 bytes
	nonceBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		nonceBytes[7-i] = byte(nonce >> (i * 8))
	}

	// Create ABI definition identical to Solidity's Call struct
	callsAbi, err := abi.JSON(strings.NewReader(`[
		{
			"name": "encodeCalls",
			"type": "function",
			"inputs": [
				{
					"name": "calls",
					"type": "tuple[]",
					"components": [
						{
							"name": "to",
							"type": "address"
						},
						{
							"name": "value",
							"type": "uint256"
						},
						{
							"name": "data",
							"type": "bytes"
						}
					]
				}
			]
		}
	]`))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Convert Go struct to Solidity struct format
	var callsForAbi []struct {
		To    common.Address
		Value *big.Int
		Data  []byte
	}

	for _, call := range calls {
		callData, err := HexToBytes(call.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid call data: %w", err)
		}

		// Handle BigInt value
		var value *big.Int
		if call.Value != nil && call.Value.Int != nil {
			value = call.Value.Int
		} else {
			value = big.NewInt(0)
		}

		callsForAbi = append(callsForAbi, struct {
			To    common.Address
			Value *big.Int
			Data  []byte
		}{
			To:    common.HexToAddress(call.To),
			Value: value,
			Data:  callData,
		})
	}

	// Perform ABI encoding (equivalent to abi.encode(_calls))
	callsEncoded, err := callsAbi.Methods["encodeCalls"].Inputs.Pack(callsForAbi)
	if err != nil {
		return nil, fmt.Errorf("failed to encode calls: %w", err)
	}

	// Implementation of abi.encodePacked
	// Equivalent to Solidity's abi.encodePacked(LEAF_ENCODING_VERSION, ONE_SIG_ID, address(this), _nonce, abi.encode(_calls))
	leafData := []byte{LeafEncodingVersion}
	leafData = append(leafData, oneSigIDBytes...) // 8 bytes
	leafData = append(leafData, addrBytes...)     // 32 bytes
	leafData = append(leafData, nonceBytes...)    // 8 bytes
	leafData = append(leafData, callsEncoded...)  // abi.encode(_calls)

	// Double hash leaf data (equivalent to Solidity's keccak256(keccak256(...)))
	firstHash := crypto.Keccak256(leafData)
	finalHash := crypto.Keccak256(firstHash)

	return finalHash, nil
}

// HexToBytes converts a hex string to bytes
func HexToBytes(hexStr string) ([]byte, error) {
	// Remove 0x prefix if present
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}

	// Convert hex to bytes
	return hex.DecodeString(hexStr)
}
