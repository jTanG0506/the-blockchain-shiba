package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const keystoreDirName = "keystore"

const ToshiAccount = "0xe5ED8C1829192380205b1E7BB5A3F44baf181d25"
const JTangAccount = "0xf70D226203FDDa745C3B160D92Ee665A71191D6a"
const QudsiiAccount = "0x7573428c0394133cC5A3FC5533b9B04241D1271E"

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}

func NewKeystoreAccount(dataDir string, password string) (common.Address, error) {
	ks := keystore.NewKeyStore(
		GetKeystoreDirPath(dataDir),
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)

	acc, err := ks.NewAccount(password)
	if err != nil {
		return common.Address{}, err
	}

	return acc.Address, nil
}

func Sign(msg []byte, privKey *ecdsa.PrivateKey) (sig []byte, err error) {
	msgHash := crypto.Keccak256(msg)

	sig, err = crypto.Sign(msgHash, privKey)
	if err != nil {
		return nil, err
	}

	if len(sig) != crypto.SignatureLength {
		return nil, fmt.Errorf(
			"wrong size for signature: got %d, expected %d",
			len(sig),
			crypto.SignatureLength,
		)
	}

	return sig, nil
}

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	msgHash := crypto.Keccak256(msg)

	publicKey, err := crypto.SigToPub(msgHash, sig)
	if err != nil {
		return nil, fmt.Errorf("unable to verify message signature. %s", err.Error())
	}

	return publicKey, nil
}
