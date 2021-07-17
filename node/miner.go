package node

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jTanG0506/go-blockchain/database"
)

type PendingBlock struct {
	parent database.Hash
	number uint64
	time   uint64
	miner  common.Address
	txs    []database.SignedTx
}

func NewPendingBlock(parent database.Hash, number uint64, miner common.Address, txs []database.SignedTx) PendingBlock {
	return PendingBlock{parent, number, uint64(time.Now().Unix()), miner, txs}
}

func Mine(ctx context.Context, pb PendingBlock) (database.Block, error) {
	if len(pb.txs) == 0 {
		return database.Block{}, fmt.Errorf("mining empty blocks is forbidden")
	}

	start := time.Now()
	nonce := uint32(0)
	var block database.Block
	var hash database.Hash

	for !database.IsBlockHashValid(hash) {
		select {
		case <-ctx.Done():
			fmt.Println("❌ Mining cancelled!")
			return database.Block{}, fmt.Errorf("mining cancelled. %s", ctx.Err())
		default:
		}

		nonce++
		if nonce == 1 || nonce%1000000 == 0 {
			fmt.Printf("⛏ Mining %d pending transactions. Attempt: %d\n", len(pb.txs), nonce)
		}

		block = database.NewBlock(pb.parent, pb.number, nonce, pb.time, pb.miner, pb.txs)
		blockHash, err := block.Hash()
		if err != nil {
			return database.Block{}, fmt.Errorf("couldn't mine block. %s", err.Error())
		}

		hash = blockHash
	}

	fmt.Printf("\nMined new Block '%x' using PoW\n", hash)
	fmt.Printf("Height: '%v'\n", block.Header.Number)
	fmt.Printf("Nonce: '%v'\n", block.Header.Nonce)
	fmt.Printf("Created: '%v'\n", block.Header.Time)
	fmt.Printf("Miner: '%v'\n", block.Header.Miner.String())
	fmt.Printf("Parent: '%v'\n", block.Header.Parent.Hex())
	fmt.Printf("Time: %s\n\n", time.Since(start))

	return block, nil
}
