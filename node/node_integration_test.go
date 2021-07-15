package node

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/fs"
	"github.com/jTanG0506/go-blockchain/wallet"
)

const MiningMaxMinutes = 30

func TestNode_Run(t *testing.T) {
	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("unexpected error when getting test data directory: %s", err)
	}

	err = fs.RemoveDir(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when removing test directory: %s", err)
	}

	n := NewNode(dataDir, "127.0.0.1", 8085, database.NewAccount(wallet.ToshiAccount), PeerNode{})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err.Error() != "http: Server closed" {
		t.Fatalf("node server expected to close after 5s but did not")
	}
}

func TestNode_Mining(t *testing.T) {
	toshi := database.NewAccount(wallet.ToshiAccount)
	jtang := database.NewAccount(wallet.JTangAccount)

	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("unexpected error when getting test data directory: %s", err)
	}

	err = fs.RemoveDir(dataDir)
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

	n := NewNode(dataDir, nodeInfo.IP, nodeInfo.Port, toshi, nodeInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*MiningMaxMinutes)

	// Add a TX in 3 seconds from now
	go func() {
		time.Sleep(time.Second * miningIntervalInSeconds / 2)
		tx := database.NewTx(toshi, jtang, 100, "")
		_ = n.AddPendingTX(tx, nodeInfo)
	}()

	// Schedule a TX in 12 seconds from now to simulate that it came in whilst
	// the first TX is being mined
	go func() {
		time.Sleep(time.Second*miningIntervalInSeconds + 2)
		tx := database.NewTx(toshi, jtang, 200, "")
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
	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("unexpected error when getting test data directory: %s", err)
	}

	err = fs.RemoveDir(dataDir)
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

	toshi := database.NewAccount(wallet.ToshiAccount)
	jtang := database.NewAccount(wallet.JTangAccount)

	n := NewNode(dataDir, nodeInfo.IP, nodeInfo.Port, toshi, nodeInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*MiningMaxMinutes)

	tx1 := database.NewTx(toshi, jtang, 100, "")
	tx2 := database.NewTx(toshi, jtang, 200, "")
	tx2Hash, _ := tx2.Hash()

	// Premine a valid block with accTwo as a miner who will receive the block
	// reward to simulate the block came on the fly from another peer
	validPreminedBlock := NewPendingBlock(database.Hash{}, 0, jtang, []database.Tx{tx1})
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
			t.Fatalf("toshi should be mining but is not")
		}

		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			t.Fatalf("failed to add block: %s", err)
		}

		// Mock that the block came from a network
		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatalf("toshi should have cancelled mining due to new synced block")
		}

		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.Hex()]
		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatalf("toshi should have cancelled mining of already mined TX")
		}

		time.Sleep(time.Second * (miningIntervalInSeconds + 2))
		if !n.isMining {
			t.Fatalf("toshi should be mining the single tx not in synced block")
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

		accOneStartBal := n.state.Balances[toshi]
		accTwoStartBal := n.state.Balances[jtang]

		// Wait until timeout reached or blocks are mined and closeNode called
		<-ctx.Done()

		accOneEndBal := n.state.Balances[toshi]
		accTwoEndBal := n.state.Balances[jtang]

		accOneExpectedEndBal := accOneStartBal - tx1.Value - tx2.Value + database.BlockReward
		accTwoExpectedEndBal := accTwoStartBal + tx1.Value + tx2.Value + database.BlockReward

		if accOneEndBal != accOneExpectedEndBal {
			t.Fatalf("expected toshi to have %d balance, not %d", accOneExpectedEndBal, accOneEndBal)
		}

		if accTwoEndBal != accTwoExpectedEndBal {
			t.Fatalf("expected jtang to have %d balance, not %d", accOneExpectedEndBal, accOneEndBal)
		}

		t.Logf("Starting toshi balance: %d", accOneStartBal)
		t.Logf("Starting jtang balance: %d", accTwoStartBal)
		t.Logf("Ending toshi balance: %d", accOneEndBal)
		t.Logf("Ending jtang balance: %d", accTwoEndBal)
	}()

	_ = n.Run(ctx)

	if n.state.LastBlock().Header.Number != 1 {
		t.Fatalf("Failed to mine the two Txs in under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatalf("Expected to have no pending TXs to mine")
	}
}

func getTestDataDirPath() (string, error) {
	return ioutil.TempDir(os.TempDir(), ".tbs_test")
}
