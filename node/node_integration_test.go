package node

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/fs"
)

const MiningMaxMinutes = 30

func TestNode_Run(t *testing.T) {
	dataDir := getTestDataDirPath()
	err := fs.RemoveDir(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when removing test directory: %s", err)
	}

	n := NewNode(dataDir, "127.0.0.1", 8081, database.NewAccount("toshi"), PeerNode{})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err.Error() != "http: Server closed" {
		t.Fatalf("node server expected to close after 5s but did not")
	}
}

func TestNode_Mining(t *testing.T) {
	dataDir := getTestDataDirPath()
	err := fs.RemoveDir(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when removing test directory: %s", err)
	}

	nodeInfo := NewPeerNode(
		"127.0.0.1",
		8081,
		false,
		database.NewAccount(""),
		true,
	)

	n := NewNode(dataDir, nodeInfo.IP, nodeInfo.Port, database.NewAccount("toshi"), nodeInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*MiningMaxMinutes)

	// Add a TX in 3 seconds from now
	go func() {
		time.Sleep(time.Second * miningIntervalInSeconds / 2)
		tx := database.NewTx("toshi", "jtang", 100, "")
		_ = n.AddPendingTX(tx, nodeInfo)
	}()

	// Schedule a TX in 12 seconds from now to simulate that it came in whilst
	// the first TX is being mined
	go func() {
		time.Sleep(time.Second*miningIntervalInSeconds + 2)
		tx := database.NewTx("toshi", "jtang", 200, "")
		_ = n.AddPendingTX(tx, nodeInfo)
	}()

	// Periodically check if the two blocks have been mined
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				if n.state.LastBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	_ = n.Run(ctx)

	if n.state.LastBlock().Header.Number != 1 {
		t.Fatalf("Failed to mine the two Txs in under 30m")
	}
}

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	dataDir := getTestDataDirPath()
	err := fs.RemoveDir(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when removing test directory: %s", err)
	}

	nodeInfo := NewPeerNode(
		"127.0.0.1",
		8081,
		false,
		database.NewAccount(""),
		true,
	)

	accOne := database.NewAccount("toshi")
	accTwo := database.NewAccount("jtang")

	n := NewNode(dataDir, nodeInfo.IP, nodeInfo.Port, accOne, nodeInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*MiningMaxMinutes)

	tx1 := database.NewTx("toshi", "jtang", 100, "")
	tx2 := database.NewTx("toshi", "jtang", 200, "")
	tx2Hash, _ := tx2.Hash()

	// Premine a valid block with accTwo as a miner who will receive the block
	// reward to simulate the block came on the fly from another peer
	validPreminedBlock := NewPendingBlock(database.Hash{}, 0, accTwo, []database.Tx{tx1})
	validSyncedBlock, err := Mine(ctx, validPreminedBlock)
	if err != nil {
		t.Fatalf("failed to produce premined / presynced block: %s", err)
	}

	go func() {
		time.Sleep(time.Second * (miningIntervalInSeconds - 2))

		err := n.AddPendingTX(tx1, nodeInfo)
		if err != nil {
			t.Fatalf("failed to add tx1: %s", err)
		}

		err = n.AddPendingTX(tx2, nodeInfo)
		if err != nil {
			t.Fatalf("failed to add tx2: %s", err)
		}
	}()

	go func() {
		time.Sleep(time.Second * (miningIntervalInSeconds + 2))
		if !n.isMining {
			t.Fatalf("accOne should be mining but is not")
		}

		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			t.Fatalf("failed to add block: %s", err)
		}

		// Mock that the block came from a network
		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatalf("accOne should have cancelled mining due to new synced block")
		}

		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.Hex()]
		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatalf("accOne should have cancelled mining of already mined TX")
		}

		time.Sleep(time.Second * (miningIntervalInSeconds + 2))
		if !n.isMining {
			t.Fatalf("accOne should be mining the single tx not in synced block")
		}
	}()

	// Periodically check if the two blocks have been mined
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				if n.state.LastBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(time.Second * 2)

		accOneStartBal := n.state.Balances[accOne]
		accTwoStartBal := n.state.Balances[accTwo]

		// Wait until timeout reached or blocks are mined and closeNode called
		<-ctx.Done()

		accOneEndBal := n.state.Balances[accOne]
		accTwoEndBal := n.state.Balances[accTwo]

		accOneExpectedEndBal := accOneStartBal - tx1.Value - tx2.Value + database.BlockReward
		accTwoExpectedEndBal := accTwoStartBal + tx1.Value + tx2.Value + database.BlockReward

		if accOneEndBal != accOneExpectedEndBal {
			t.Fatalf("expected accOne to have %d balance, not %d", accOneExpectedEndBal, accOneEndBal)
		}

		if accTwoEndBal != accTwoExpectedEndBal {
			t.Fatalf("expected accOne to have %d balance, not %d", accOneExpectedEndBal, accOneEndBal)
		}

		t.Logf("Starting accOne balance: %d", accOneStartBal)
		t.Logf("Starting accTwo balance: %d", accTwoStartBal)
		t.Logf("Ending accOne balance: %d", accOneEndBal)
		t.Logf("Ending accTwo balance: %d", accTwoEndBal)
	}()

	_ = n.Run(ctx)

	if n.state.LastBlock().Header.Number != 1 {
		t.Fatalf("Failed to mine the two Txs in under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatalf("Expected to have no pending TXs to mine")
	}
}

func getTestDataDirPath() string {
	return filepath.Join(os.TempDir(), ".tbs_test")
}
