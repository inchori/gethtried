# gethtried

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A CLI tool for visualizing Ethereum's Merkle Patricia Tries from an archive node RPC.

## Features

- **State Trie Visualization**: Visualize account proofs using `eth_getProof`
- **Storage Trie Visualization**: Inspect contract storage slot proofs
- **Transaction Trie Verification**: Verify and display transaction tries
- **Receipt Trie Verification**: Verify and display receipt tries

## Requirements

- Go 1.22+
- Geth-compatible archive node RPC endpoint

## Installation

```bash
git clone https://github.com/inchori/gethtried.git
cd gethtried
make build
```

## Usage

### State Trie (Account Proof)

```bash
./build/gethtried state \
  --rpc-url https://your-archive-node.com \
  --block-height 18000000 \
  --account-address 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

### Storage Trie

```bash
./build/gethtried storage \
  --rpc-url https://your-archive-node.com \
  --block-height 18000000 \
  --account-address 0xA0b86a33E6441b8C4505A3b5a4F5B6D4D1c8f8c8 \
  --slot 0
```

### Transaction Trie

```bash
./build/gethtried tx \
  --rpc-url https://your-archive-node.com \
  --block-height 18000000
```

### Receipt Trie

```bash
./build/gethtried receipt \
  --rpc-url https://your-archive-node.com \
  --block-height 18000000
```

## Commands

| Command | Description | Required Flags |
|---------|-------------|---------------|
| `state` | Visualize account state proof | `--block-height`, `--account-address` |
| `storage` | Visualize storage slot proof | `--block-height`, `--account-address`, `--slot` |
| `tx` | Verify transaction trie | `--block-height` |
| `receipt` | Verify receipt trie | `--block-height` |

## Example Output

```
--- Logical Trie Path Visualization ---
Target Path: a7f9365b9c4b8b8e8f5c8a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9

├── KEY: 0x8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b
│   Type: Branch
│   - Has Value: false
│   -> Branching: Following path nibble 'a' (index 10)
│   ├── KEY: 0x7f9365b9c4b8b8e8f5c8a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9
│   │   Type: Extension
│   │   - Shared Path: '7f9365b9c4b8b8e8f5c8a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8'
│   │   -> Following Extension Node...
│   │   └── Leaf Reached. Final Value:
│   │       - Nonce:       42
│   │       - Balance:     1.500000 ETH
│   │       - StorageRoot: 0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
│   │       - CodeHash:    0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
```

## How It Works

The tool fetches Merkle proofs via RPC, parses RLP-encoded trie nodes, and visualizes the path traversal through the trie structure. It correctly handles both 32-byte hash references and inline RLP nodes.

## License

MIT License - see LICENSE file for details.