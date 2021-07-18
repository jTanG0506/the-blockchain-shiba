package node

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/wallet"
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
	minerPrivateKey, _, miner, err := generateKey()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %s", err.Error())
	}

	pendingBlock, err := createRandomPendingBlock(minerPrivateKey, miner)
	if err != nil {
		t.Fatalf("Failed to create random pending block: %s", err.Error())
	}

	ctx := context.Background()

	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatalf("Failed to mine block: %s", err.Error())
	}

	minedBlockHash, err := minedBlock.Hash()
	if err != nil {
		t.Fatalf("Failed to retrieve hash of mined block: %s", err.Error())
	}

	if !database.IsBlockHashValid(minedBlockHash) {
		t.Fatalf("Mined block has invalid block hash: %s", err.Error())
	}

	if minedBlock.Header.Miner.String() != miner.String() {
		t.Fatalf("Mined blocked miner was %s, expected %s", minedBlock.Header.Miner, miner)
	}
}

func TestMineWithTimeout(t *testing.T) {
	minerPrivateKey, _, miner, err := generateKey()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %s", err.Error())
	}

	pendingBlock, err := createRandomPendingBlock(minerPrivateKey, miner)
	if err != nil {
		t.Fatalf("Failed to create random pending block: %s", err.Error())
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Microsecond*100)

	_, err = Mine(ctx, pendingBlock)
	if err == nil {
		t.Fatalf("Expected mine to give error from timeout")
	}
}

func generateKey() (*ecdsa.PrivateKey, ecdsa.PublicKey, common.Address, error) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, ecdsa.PublicKey{}, common.Address{}, err
	}

	publicKey := privateKey.PublicKey
	publicKeyBytes := elliptic.Marshal(crypto.S256(), publicKey.X, publicKey.Y)
	publicKeyBytesHash := crypto.Keccak256(publicKeyBytes[1:])
	account := common.BytesToAddress(publicKeyBytesHash[12:])

	return privateKey, publicKey, account, nil
}

func createRandomPendingBlock(privateKey *ecdsa.PrivateKey, acc common.Address) (PendingBlock, error) {
	tx := database.NewTx(acc, database.NewAccount(wallet.ToshiAccount), 100, "test")
	signedTx, err := wallet.SignTx(tx, privateKey)
	if err != nil {
		return PendingBlock{}, err
	}

	return NewPendingBlock(
		database.Hash{},
		0,
		acc,
		[]database.SignedTx{signedTx},
	), nil
}
