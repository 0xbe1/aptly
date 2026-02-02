# Aptos Node API Coverage

This document tracks the implementation status of Aptos Node API endpoints in the `apt` CLI.

## Implemented Endpoints

| Endpoint | Command | Status |
|----------|---------|--------|
| GET `/` | `apt node ledger` | Done |
| GET `/spec.json` | `apt node spec` | Done |
| GET `/-/healthy` | `apt node health` | Done |
| GET `/info` | `apt node info` | Done |
| GET `/accounts/{address}` | `apt account <address>` | Done |
| GET `/accounts/{address}/resources` | `apt account resources <address>` | Done |
| GET `/accounts/{address}/resource/{resource_type}` | `apt account resource <address> <type>` | Done |
| GET `/accounts/{address}/modules` | `apt account modules <address>` | Done |
| GET `/accounts/{address}/module/{module_name}` | `apt account module <address> <name>` | Done |
| GET `/accounts/{address}/balance/{asset_type}` | `apt account balance <address> [asset]` | Done |
| GET `/accounts/{address}/transactions` | `apt account txs <address>` | Done |
| GET `/accounts/{address}/events/{creation_number}` | `apt events <address> <creation_number>` | Done |
| GET `/blocks/by_height/{block_height}` | `apt block <height>` | Done |
| GET `/blocks/by_version/{version}` | `apt block by-version <version>` | Done |
| GET `/transactions` | `apt tx list` | Done |
| GET `/transactions/by_version/{version}` | `apt tx <version>` | Done |
| GET `/transactions/by_hash/{hash}` | `apt tx <hash>` | Done |
| POST `/transactions` | `apt tx submit` | Done |
| POST `/transactions/encode_submission` | `apt tx encode` | Done |
| POST `/transactions/simulate` | `apt tx simulate` | Done |
| POST `/tables/{table_handle}/item` | `apt table item <handle>` | Done |
| POST `/view` | `apt view <function>` | Done |

## Not Implemented

| Endpoint | Reason |
|----------|--------|
| POST `/transactions/batch` | Batch submission rarely needed for agents |
| GET `/estimate_gas_price` | Can be derived from recent transactions |

## Innovative Features (Beyond API)

These commands provide value-added analysis not available in the raw API:

| Command | Description |
|---------|-------------|
| `apt tx balance-change` | Analyzes FungibleStore changes to show net balance impact |
| `apt tx transfers` | Extracts Withdraw/Deposit events into structured transfers |
| `apt tx graph` | Pairs withdraw→deposit events to visualize asset flow |
| `apt tx trace` | Integrates with Sentio API for call trace analysis |
| Stdin piping | All tx commands support `apt tx <ver> \| apt tx transfers` |

## Command Reference

### Node Commands

```bash
# Get current ledger info
apt node ledger

# Get OpenAPI specification
apt node spec

# Check node health
apt node health

# Get node info
apt node info
```

### Account Commands

```bash
# Get account info (auth key, sequence number)
apt account 0x1

# List all resources
apt account resources 0x1

# Get specific resource
apt account resource 0x1 0x1::account::Account
apt account resource 0x1 "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>"

# Get balance (defaults to APT)
apt account balance 0x1
apt account balance 0x1 0x1::aptos_coin::AptosCoin

# List modules
apt account modules 0x1

# Get specific module
apt account module 0x1 coin

# List account transactions
apt account txs 0x1 --limit 10
```

### Block Commands

```bash
# Get block by height
apt block 1000000
apt block 1000000 --with-transactions

# Get block by transaction version
apt block by-version 2658869495
```

### Events Command

```bash
# Get events by creation number
apt events 0x1 0 --limit 10
```

### Table Command

```bash
# Get table item
apt table item 0x1b854694ae746cdbd8d44186ca4929b2b337df21d1c74633be19b2710552fdca \
  --key-type address \
  --value-type "0x1::staking_contract::StakingContract" \
  --key '"0x1"'
```

### View Command

```bash
# Execute view function
apt view 0x1::coin::balance --type-args 0x1::aptos_coin::AptosCoin --args '"0x1"'
apt view 0x1::account::exists_at --args '"0x1"'
```

### Transaction Commands

```bash
# List recent transactions
apt tx list --limit 10

# View transaction by version or hash
apt tx 2658869495
apt tx 0x123abc...

# Analyze transaction
apt tx 2658869495 | apt tx balance-change
apt tx 2658869495 | apt tx transfers
apt tx 2658869495 | apt tx graph

# Simulate transaction
cat payload.json | apt tx simulate 0x1

# Submit/encode workflow
cat unsigned_tx.json | apt tx encode
# Sign externally, then:
cat signed_tx.json | apt tx submit
```
