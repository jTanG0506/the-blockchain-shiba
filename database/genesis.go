package database

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/common"
)

var genesisJson = `
{
  "genesis_time": "2020-06-07T00:00:00.000000000Z",
  "chain_id": "the-blockchain-shiba-ledger",
  "balances": {
    "0xe5ED8C1829192380205b1E7BB5A3F44baf181d25": 1000000
  }
}`

type Genesis struct {
	Balances map[common.Address]uint `json:"balances"`
}

func loadGenesis(path string) (Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var loadedGenesis Genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return Genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string, genesis []byte) error {
	return ioutil.WriteFile(path, genesis, 0644)
}
