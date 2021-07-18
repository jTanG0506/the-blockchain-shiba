package database

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewAccount(value string) common.Address {
	return common.HexToAddress(value)
}

type Tx struct {
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Value uint           `json:"value"`
	Nonce uint           `json:"nonce"`
	Data  string         `json:"data"`
	Time  uint64         `json:"time"`
}

type SignedTx struct {
	Tx
	Sig []byte `json:"signature"`
}

func NewTx(from common.Address, to common.Address, value, nonce uint, data string) Tx {
	return Tx{from, to, value, nonce, data, uint64(time.Now().Unix())}
}

func NewSignedTx(tx Tx, sig []byte) SignedTx {
	return SignedTx{tx, sig}
}

func (t Tx) IsReward() bool {
	return t.Data == "reward"
}

func (t Tx) Hash() (Hash, error) {
	txJson, err := t.Encode()
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(txJson), nil
}

func (t Tx) Encode() ([]byte, error) {
	return json.Marshal(t)
}

func (t SignedTx) Hash() (Hash, error) {
	txJson, err := t.Encode()
	if err != nil {
		return Hash{}, err
	}

	return sha256.Sum256(txJson), nil
}

func (t SignedTx) IsSigAuthentic() (bool, error) {
	txHash, err := t.Tx.Hash()
	if err != nil {
		return false, err
	}

	recoveredPubKey, err := crypto.SigToPub(txHash[:], t.Sig)
	if err != nil {
		return false, err
	}

	recoveredPubKeyBytes := elliptic.Marshal(
		crypto.S256(),
		recoveredPubKey.X,
		recoveredPubKey.Y,
	)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	return recoveredAccount.Hex() == t.From.Hex(), nil
}
