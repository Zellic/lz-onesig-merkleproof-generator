package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
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

// EncodeLeaf encodes a transaction as a leaf according to OneSig spec
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

		callsForAbi = append(callsForAbi, struct {
			To    common.Address
			Value *big.Int
			Data  []byte
		}{
			To:    common.HexToAddress(call.To),
			Value: call.Value,
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
