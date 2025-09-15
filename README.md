# gethtried

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A Go-based CLI tool for visualizing Ethereum's Merkle Patricia Tries directly from an archive node RPC.

## Features

-   [x] **State Trie Visualization:** Visualize the Merkle proof path for any account (`eth_getProof`) at any historical block height.
-   [ ] **Storage Trie Visualization:** (Upcoming) Inspect the storage trie proof for a specific contract address and slot.
-   [ ] **Transaction Trie Visualization:** (Upcoming) Render the transaction trie for a given block.
-   [ ] **Receipts Trie Visualization:** (Upcoming) Render the receipts trie for a given block.

## Requirements

- Go 1.25+
- Geth-compatible Archive Node RPC endpoint (historical state requires archive)

## Installation

```bash
git clone https://github.com/inchori/gethtried.git
cd gethtried
make build
```

## Usage
Currently, only the state command is implemented.

```
./build/gethtried state --rpc-url <YOUR_ARCHIVE_NODE_URL> --block-height <BLOCK_NUMBER> --account-address <ACCOUNT_ADDRESS>
```

Arguments:

- `--rpc-url` (default: `http://localhost:8545`): Geth Archive Node RPC URL
- `--block-height` (required): Historical block number to query
- `--account-address` (required): EOA or contract address to generate proof for

### Help

```bash
./build/gethtried --help
./build/gethtried state --help
```

## License
This project is licensed under the MIT License. See the LICENSE file for details.