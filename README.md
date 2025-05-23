# OneSig Merkle Root Generator

OneSig Merkle Root Generator is a CLI tool for generating Merkle roots for transaction batches according to the LayerZero OneSig specification.

## Usage

```bash
# Basic usage
./merkle-cli --onesig-id [ONESIG_ID] --batch-file [PATH_TO_BATCH_FILE]

# Specify a contract address
./merkle-cli --onesig-id [ONESIG_ID] --contract-addr [CONTRACT_ADDRESS] --batch-file [PATH_TO_BATCH_FILE]

# Example
./merkle-cli --onesig-id 1 --batch-file ./examples/sample-batch.json

# Detailed output (including proofs)
./merkle-cli --onesig-id 1 --batch-file ./examples/sample-batch.json --verbose
```

### Options

- `--onesig-id`, `-o`: OneSig ID (typically Chain ID)
- `--contract-addr`, `-c`: OneSig contract address (defaults to 0xdEaD if not provided)
- `--batch-file`, `-f`: Path to JSON file defining the transaction batch
- `--verbose`, `-v`: Show detailed output including Merkle proofs

## Transaction Batch JSON Format

### Recommended Format (Group-based)

```json
{
  "groups": [
    {
      "nonce": 0,
      "calls": [
        {
          "to": "0xfEdcBA9876543210FedCBa9876543210fEdCBa98",
          "value": 500,
          "data": "0x"
        },
        {
          "to": "0xfEdcBA9876543210FedCBa9876543210fEdCBa98",
          "value": 1000000000000000000,
          "data": "0x"
        }
      ]
    }
  ]
}
```

Each group contains the following fields:
- `nonce`: Nonce value (integer) - all calls with the same nonce are encoded as a single leaf
- `calls`: List of calls
  - `to`: Target address (hexadecimal string)
  - `value`: Value to send (hexadecimal or decimal string)
  - `data`: Call data (hexadecimal string)

