package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io/ioutil"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/fs"
)

const testKsPassword = "password"

func TestSign(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatalf("unable to generate keypair. %s", err.Error())
	}

	pubKey := privKey.PublicKey
	pubKeyBytes := elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	pubKeyBytesHash := crypto.Keccak256(pubKeyBytes[1:])

	account := common.BytesToAddress(pubKeyBytesHash[12:])
	msg := []byte("bitcoin to the moon")

	sig, err := Sign(msg, privKey)
	if err != nil {
		t.Fatalf("unable to sign message with private key. %s", err.Error())
	}

	recoveredPubKey, err := Verify(msg, sig)
	if err != nil {
		t.Fatalf("unable to recover public key from signed message. %s", err.Error())
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	if account.Hex() != recoveredAccount.Hex() {
		t.Fatalf("msg was signed by %s but signature recover produced account %s", account.Hex(), recoveredAccount.Hex())
	}
}

func TestSignTxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wallet_test")
	if err != nil {
		t.Fatalf("unable to create temporary directory. %s", err.Error())
	}
	defer fs.RemoveDir(tmpDir)

	toshi, err := NewKeystoreAccount(tmpDir, testKsPassword)
	if err != nil {
		t.Errorf("unable to create keystore account. %s", err.Error())
	}

	jtang, err := NewKeystoreAccount(tmpDir, testKsPassword)
	if err != nil {
		t.Errorf("unable to create keystore account. %s", err.Error())
	}

	tx := database.NewTx(toshi, jtang, 100, 1, "")
	signedTx, err := SignTxWithKeystoreAccount(
		tx,
		toshi,
		testKsPassword,
		GetKeystoreDirPath(tmpDir),
	)
	if err != nil {
		t.Fatalf("unable to sign transaction with private key. %s", err.Error())
	}

	ok, err := signedTx.IsSigAuthentic()
	if err != nil {
		t.Fatalf("unable to determine whether signature is authentic. %s", err.Error())
	}

	if !ok {
		t.Fatalf("signature on transaction is not authentic. %s", err.Error())
	}
}

func TestSignedForgedTxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wallet_test")
	if err != nil {
		t.Fatalf("unable to create temporary directory. %s", err.Error())
	}
	defer fs.RemoveDir(tmpDir)

	hacker, err := NewKeystoreAccount(tmpDir, testKsPassword)
	if err != nil {
		t.Errorf("unable to create keystore account. %s", err.Error())
	}

	toshi, err := NewKeystoreAccount(tmpDir, testKsPassword)
	if err != nil {
		t.Errorf("unable to create keystore account. %s", err.Error())
	}

	forgedTx := database.NewTx(toshi, hacker, 100, 1, "")
	signedTx, err := SignTxWithKeystoreAccount(
		forgedTx,
		hacker,
		testKsPassword,
		GetKeystoreDirPath(tmpDir),
	)
	if err != nil {
		t.Fatalf("unable to sign transaction with private key. %s", err.Error())
	}

	ok, err := signedTx.IsSigAuthentic()
	if err != nil {
		t.Fatalf("unable to determine whether signature is authentic. %s", err.Error())
	}

	if ok {
		t.Fatalf("signature on transaction should not be authentic. %s", err.Error())
	}
}
