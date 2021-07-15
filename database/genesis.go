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

type genesis struct {
	Balances map[common.Address]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(genesisJson), 0644)
}
