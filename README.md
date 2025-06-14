# OneSig Merkle Tree Generator

A CLI tool for generating Merkle trees and proofs for OneSig transaction batches according to the LayerZero OneSig specification.

## Features

- **Leaf Encoding**: Process JSON input to encode leaves according to specified versions
- **Merkle Tree Generation**: Generate Merkle trees with proofs for each leaf
- **Standalone Merkle Logic**: Generate Merkle trees from pre-encoded leaves
- **Input Validation**: Ensure no duplicate nonces within the same oneSigId
- **Flexible Options**: Support for sortedPairs and sortLeaves options
- **Multiple Input Formats**: Support for JSON files and comma-separated values

## Installation

```bash
go build -o merkle-cli
```

## Usage

The tool supports two main operations:

### 1. Encode Command

Process JSON input to encode leaves and generate merkle tree with proofs:

```bash
./merkle-cli encode --file-path input.json --leafEncodingVersion 1
```

**Flags:**
- `--file-path, -f`: Path to the JSON file containing the leaves (required)
- `--leafEncodingVersion, -v`: Encoding version to use for the leaves (default: 1)

### 2. Merkle Command

Generate merkle tree from pre-encoded leaves:

```bash
# From file
./merkle-cli merkle --encodedInput encoded.json --sortedPairs true

# From comma-separated values
./merkle-cli merkle --encodedInput "0xabc...,0xdef..." --sortedPairs false
```

**Flags:**
- `--encodedInput, -i`: File path to JSON array of encoded leaves or comma-separated array (required)
- `--sortedPairs, -s`: Use sorted pairs when building the Merkle Tree (default: true)
- `--sortLeaves, -l`: Sort leaves before building the tree (default: false)

## Input Format

### For Encode Command

```json
{
  "leaves": [
    {
      "nonce": "7",
      "oneSigId": "40231",
      "targetOneSigAddress": "0x9f7cf878377c1ead0bb4ee9bf76955f633f70c62",
      "calls": [
        {
          "to": "0x9f7cf878377c1ead0bb4ee9bf76955f633f70c62",
          "value": "0",
          "data": "0x"
        }
      ]
    }
  ]
}
```

### For Merkle Command

```json
{
  "encodedLeaves": [
    "0x29c07bb0d5d5fb6c40a87b4e84726aa2a8f28918cd839ce5850a3585de000307",
    "0x78f2743fb1103d8f5033829224d8b918e32af0478fa85c158045f16dda4dc05c"
  ]
}
```

## Output Format

Both commands produce structured JSON output:

```json
{
  "merkleRoot": "0xcd43b77dcfe46936536e7c5762bde4b8ef32dacaa35c6617212c39a154480cd9",
  "proofs": [
    {
      "leaf": "0x29c07bb0d5d5fb6c40a87b4e84726aa2a8f28918cd839ce5850a3585de000307",
      "nonce": "9",
      "oneSigId": "40102",
      "targetOneSigAddress": "0xc3f959962e3d2e57d12c85dd35075e5f04f19be7",
      "proof": [
        "0x78f2743fb1103d8f5033829224d8b918e32af0478fa85c158045f16dda4dc05c",
        "0xeba30326834a60ebbcaa66af6d3ee6439cac0a22c43d2f0d7f056d18418a07ca"
      ]
    }
  ]
}
```

## Validation Rules

- No duplicate nonces are allowed for the same `oneSigId`
- Duplicate nonces are allowed across different `oneSigIds`
- All required fields must be present and valid
- Value fields support both string and number formats

## Examples

### Basic Usage

```bash
# Generate merkle tree from JSON input
./merkle-cli encode --file-path examples/new-format-input.json

# Generate merkle tree from encoded leaves with sorted pairs
./merkle-cli merkle --encodedInput examples/encoded-leaves.json --sortedPairs true

# Generate merkle tree from comma-separated encoded leaves
./merkle-cli merkle --encodedInput "0x29c07bb0d5d5fb6c40a87b4e84726aa2a8f28918cd839ce5850a3585de000307,0x78f2743fb1103d8f5033829224d8b918e32af0478fa85c158045f16dda4dc05c"
```

### Advanced Options

```bash
# Use specific encoding version
./merkle-cli encode --file-path input.json --leafEncodingVersion 1

# Sort leaves before building tree
./merkle-cli merkle --encodedInput encoded.json --sortLeaves true --sortedPairs false
```

## Example Files

The `examples/` directory contains sample input files:

- `new-format-input.json`: Example input for encode command
- `encoded-leaves.json`: Example input for merkle command
- `sample-batch.json`: Legacy format (for backward compatibility)

## Development

### Project Structure

```
├── cmd/           # CLI commands
├── examples/      # Example input files
├── merkle/        # Merkle tree implementation
├── models/        # Data models and types
├── utils/         # Utility functions
├── main.go        # Entry point
└── README.md      # This file
```

### Building

```bash
go build -o merkle-cli
```

### Testing

```bash
go test ./...
```

## Compatibility

This tool follows the MerkleTreeJs interface for merkle tree operations and is compatible with the LayerZero OneSig specification.

