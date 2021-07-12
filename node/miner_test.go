package node

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/jTanG0506/go-blockchain/database"
)

func TestValidBlockHash(t *testing.T) {
	hexHash := "0000009fd186c1dbc756317bcd5711442effca7aaa6c9e5c4c59670c5de5a7ad"
	var hash = database.Hash{}
	hex.Decode(hash[:], []byte(hexHash))

	isValid := database.IsBlockHashValid(hash)
	if !isValid {
		t.Fatalf("Hash '%s' with 6 zeros should be valid", hexHash)
	}
}

func TestInvalidBlockHash(t *testing.T) {
	hexHash := "7e2ddf9fd186c1dbc756317bcd5711442effca7aaa6c9e5c4c59670c5de5a7ad"
	var hash = database.Hash{}
	hex.Decode(hash[:], []byte(hexHash))

	isValid := database.IsBlockHashValid(hash)
	if isValid {
		t.Fatalf("Hash '%s' without 6 zeros should be invalid", hexHash)
	}
}

func TestMine(t *testing.T) {
	pendingBlock := createRandomPendingBlock()
	ctx := context.Background()

	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatalf("Failed to mine block: %x", err)
	}

	minedBlockHash, err := minedBlock.Hash()
	if err != nil {
		t.Fatalf("Failed to retrieve hash of mined block: %x", err)
	}

	if !database.IsBlockHashValid(minedBlockHash) {
		t.Fatalf("Mined block has invalid block hash: %x", err)
	}
}

func TestMineWithTimeout(t *testing.T) {
	pendingBlock := createRandomPendingBlock()
	ctx, _ := context.WithTimeout(context.Background(), time.Microsecond*100)

	_, err := Mine(ctx, pendingBlock)
	if err == nil {
		t.Fatalf("Expected mine to give error from timeout")
	}
}

func createRandomPendingBlock() PendingBlock {
	return NewPendingBlock(
		database.Hash{},
		0,
		database.Account("toshi"),
		[]database.Tx{
			database.NewTx("toshi", "jtang", 1000, ""),
			database.NewTx("toshi", "toshi", 10, "reward"),
		},
	)
}
