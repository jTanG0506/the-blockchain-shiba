package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

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
