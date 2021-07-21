# The Blockchain Shiba

- ❓ A peer-to-peer, autonomous blockchain system in Go
- ⛽️ Fixed gas fee per transaction and block reward
- 🔒 Uses `ethereum/go-ethereum` for wallet and keystores

# Usage

## Installation

```
go install ./cmd/...
```

## CLI

### Show available commands and flags

```
tbs help
```

### Show available run settings

```
~ tbs run --help
Launches the TBS node and its HTTP API

Usage:
  tbs run [flags]

Flags:
      --bootstrap-account string   default bootstrap account to interconnect peers (default "0xe5ED8C1829192380205b1E7BB5A3F44baf181d25")
      --bootstrap-ip string        default bootstrap server to interconnect peers (default "127.0.0.1")
      --bootstrap-port uint        default bootstrap server port to interconnect peers (default 8080)
      --datadir string             Absolute path to the node data dit where the DB will be stored
  -h, --help                       help for run
      --ip string                  exposed IP for communication with peers (default "127.0.0.1")
      --miner string               miner account of this node to receive block rewards (default "0x0000000000000000000000000000000000000000")
      --port uint                  exposed HTTP port for communication with peers (default 8080)
```

### Notes

- The genesis block can be found at `database/genesis.json`
- To modify the network difficulty, modify the `IsBlockHashValid` function in `database/state.go`

## Tests

Run all tests with verbosity but one at a time (without timeout as mining takes time), to avoid ports collisions:

```
go test -v -p=1 -timeout=0 ./...
```

Run an individual test:

```
go test -timeout=0 ./node -test.v -test.run ^TestNode_Mining$
```

Note: Majority are integration tests and will take time to run, due to mining. Some accounts have been added to the repository as test accounts.

## Who is Toshi?

[Toshi](https://www.instagram.com/shiba.the.toshi/) is my Shiba Inu, named after the pseudonymous person who developed Bitcoin, Satoshi Nakamoto.
