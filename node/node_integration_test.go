package node

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/fs"
	"github.com/jTanG0506/go-blockchain/wallet"
)

const MiningMaxMinutes = 30

const testKsToshiAccount = "0xe5ED8C1829192380205b1E7BB5A3F44baf181d25"
const testKsJTangAccount = "0xf70D226203FDDa745C3B160D92Ee665A71191D6a"
const testKsQudsiiAccount = "0x7573428c0394133cC5A3FC5533b9B04241D1271E"
const testKsToshiFile = "toshi--e5ed8c1829192380205b1e7bb5a3f44baf181d25"
const testKsJTangFile = "jtang--f70d226203fdda745c3b160d92ee665a71191d6a"
const testKsQudsiiFile = "qudsii--7573428c0394133cc5a3fc5533b9b04241d1271e"
const testKsToshiPwd = "toshi"
const testKsJTangPwd = "jtang"
const testKsQudsiiPwd = "qudsii"

func TestNode_Run(t *testing.T) {
	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("unexpected error when getting test data directory: %s", err)
	}

	err = fs.RemoveDir(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when removing test directory: %s", err)
	}

	n := NewNode(dataDir, "127.0.0.1", 8085, database.NewAccount(DefaultMiner), PeerNode{})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err.Error() != "http: Server closed" {
		t.Fatalf("node server expected to close after 5s but did not")
	}
}

func TestNode_Mining(t *testing.T) {
	toshi := database.NewAccount(testKsToshiAccount)
	jtang := database.NewAccount(testKsJTangAccount)

	genesisBalances := make(map[common.Address]uint)
	genesisBalances[toshi] = 1000000
	genesis := database.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		t.Fatalf("unexpected error when creating test genesis: %s", err.Error())
	}

	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("unexpected error when getting test data directory: %s", err)
	}

	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	if err != nil {
		t.Fatalf("unexpected error when initialising dataDir: %s", err)
	}
	defer fs.RemoveDir(dataDir)

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when copying keystore files to test directory: %s", err)
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

		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx,
			toshi,
			testKsToshiPwd,
			wallet.GetKeystoreDirPath(dataDir),
		)
		if err != nil {
			t.Error(err)
			return
		}
		_ = n.AddPendingTX(signedTx, nodeInfo)
	}()

	// Schedule a TX in 12 seconds from now to simulate that it came in whilst
	// the first TX is being mined
	go func() {
		time.Sleep(time.Second*miningIntervalInSeconds + 2)
		tx := database.NewTx(toshi, jtang, 200, "")

		signedTx, err := wallet.SignTxWithKeystoreAccount(
			tx,
			toshi,
			testKsToshiPwd,
			wallet.GetKeystoreDirPath(dataDir),
		)
		if err != nil {
			t.Error(err)
			return
		}
		_ = n.AddPendingTX(signedTx, nodeInfo)
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
	toshi := database.NewAccount(testKsToshiAccount)
	jtang := database.NewAccount(testKsJTangAccount)

	genesisBalances := make(map[common.Address]uint)
	genesisBalances[toshi] = 1000000
	genesis := database.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		t.Fatalf("unexpected error when creating test genesis: %s", err.Error())
	}

	dataDir, err := getTestDataDirPath()
	if err != nil {
		t.Fatalf("unexpected error when getting test data directory: %s", err)
	}

	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	if err != nil {
		t.Fatalf("unexpected error when initialising dataDir: %s", err)
	}
	defer fs.RemoveDir(dataDir)

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		t.Fatalf("unexpected error when copying keystore files to test directory: %s", err)
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

	tx1 := database.NewTx(toshi, jtang, 100, "")
	tx2 := database.NewTx(toshi, jtang, 200, "")

	signedTx1, err := wallet.SignTxWithKeystoreAccount(tx1, toshi, testKsToshiAccount, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Fatalf("unable to sign tx1 with keystore account: %s", err.Error())
	}

	signedTx2, err := wallet.SignTxWithKeystoreAccount(tx2, toshi, testKsToshiAccount, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Fatalf("unable to sign tx1 with keystore account: %s", err.Error())
	}

	tx2Hash, err := signedTx2.Hash()
	if err != nil {
		t.Fatalf("unexpected error when hashing signedTx2: %s", err.Error())
	}

	// Premine a valid block with accTwo as a miner who will receive the block
	// reward to simulate the block came on the fly from another peer
	validPreminedBlock := NewPendingBlock(database.Hash{}, 0, jtang, []database.SignedTx{signedTx1})
	validSyncedBlock, err := Mine(ctx, validPreminedBlock)
	if err != nil {
		t.Fatalf("failed to produce premined / presynced block: %s", err)
	}

	go func() {
		time.Sleep(time.Second * (miningIntervalInSeconds - 2))

		err := n.AddPendingTX(signedTx1, nodeInfo)
		if err != nil {
			t.Fatalf("failed to add tx1: %s", err)
		}

		err = n.AddPendingTX(signedTx2, nodeInfo)
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

// Copy the pregenerated, committed keystore files into a given dirPath for testing
func copyKeystoreFilesIntoTestDataDirPath(dataDir string) error {
	ksDir := filepath.Join(wallet.GetKeystoreDirPath(dataDir))
	err := os.Mkdir(ksDir, 0777)
	if err != nil {
		return err
	}

	// Copy Toshi's Account
	toshiSrcKs, err := os.Open(testKsToshiFile)
	if err != nil {
		return err
	}
	defer toshiSrcKs.Close()

	toshiDestKs, err := os.Create(filepath.Join(ksDir, testKsToshiFile))
	if err != nil {
		return err
	}
	defer toshiDestKs.Close()

	_, err = io.Copy(toshiSrcKs, toshiDestKs)
	if err != nil {
		return err
	}

	// Copy jTanG's Account
	jtangSrcKs, err := os.Open(testKsJTangFile)
	if err != nil {
		return err
	}
	defer jtangSrcKs.Close()

	jtangDestKs, err := os.Create(filepath.Join(ksDir, testKsJTangFile))
	if err != nil {
		return err
	}
	defer jtangDestKs.Close()

	_, err = io.Copy(toshiSrcKs, toshiDestKs)
	if err != nil {
		return err
	}

	// Copy Qudsii's Account
	qudsiiSrcKs, err := os.Open(testKsJTangFile)
	if err != nil {
		return err
	}
	defer qudsiiSrcKs.Close()

	qudsiiDestKs, err := os.Create(filepath.Join(ksDir, testKsQudsiiFile))
	if err != nil {
		return err
	}
	defer qudsiiDestKs.Close()

	_, err = io.Copy(toshiSrcKs, toshiDestKs)
	if err != nil {
		return err
	}

	return nil
}
