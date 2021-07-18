package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

type State struct {
	Balances        map[common.Address]uint
	AccountsToNonce map[common.Address]uint

	dbFile          *os.File
	lastBlock       Block
	lastBlockHash   Hash
	hasGenesisBlock bool
}

func NewStateFromDisk(dataDir string) (*State, error) {
	err := InitDataDirIfNotExists(dataDir, []byte(genesisJson))
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[common.Address]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	accountToNonce := make(map[common.Address]uint)
	blocksFilePath := getBlocksDbFilePath(dataDir)
	blocks, err := os.OpenFile(blocksFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(blocks)
	state := &State{balances, accountToNonce, blocks, Block{}, Hash{}, false}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()

		if len(blockFsJson) == 0 {
			break
		}

		var blockFs BlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		err = applyBlock(blockFs.Value, state)
		if err != nil {
			return nil, err
		}

		state.lastBlock = blockFs.Value
		state.lastBlockHash = blockFs.Key
		state.hasGenesisBlock = true
	}

	return state, nil
}

func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.LastBlock().Header.Number + 1
}

func (s *State) LastBlock() Block {
	return s.lastBlock
}

func (s *State) LatestBlockHash() Hash {
	return s.lastBlockHash
}

func (s *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		_, err := s.AddBlock(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *State) AddBlock(b Block) (Hash, error) {
	tempState := s.copy()

	err := applyBlock(b, &tempState)
	if err != nil {
		return Hash{}, err
	}

	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{Key: blockHash, Value: b}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new block to disk:\n")
	fmt.Printf("%s\n", blockFsJson)

	_, err = s.dbFile.Write(append(blockFsJson, '\n'))
	if err != nil {
		return Hash{}, err
	}

	s.Balances = tempState.Balances
	s.AccountsToNonce = tempState.AccountsToNonce
	s.lastBlockHash = blockHash
	s.lastBlock = b
	s.hasGenesisBlock = true

	return blockHash, nil
}

func applyBlock(b Block, s *State) error {
	expectedNextBlockNumber := s.lastBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != expectedNextBlockNumber {
		return fmt.Errorf("next expected block must have number '%d' not '%d'", expectedNextBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.lastBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.lastBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.lastBlockHash, b.Header.Parent)
	}

	hash, err := b.Hash()
	if err != nil {
		return err
	}

	if !IsBlockHashValid(hash) {
		return fmt.Errorf("invalid block hash %x", hash)
	}

	err = applyTXs(b.TXs, s)
	if err != nil {
		return err
	}

	s.Balances[b.Header.Miner] += BlockReward
	return nil
}

func applyTXs(txs []SignedTx, s *State) error {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Time < txs[j].Time
	})

	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyTx(tx SignedTx, s *State) error {
	ok, err := tx.IsSigAuthentic()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("wrong Tx, sender '%s' is forged", tx.From.String())
	}

	expectedNonce := s.GetNextAccountNonce(tx.From)
	if tx.Nonce != expectedNonce {
		fmt.Errorf("wrong Tx, sender '%s' next nonce must be '%d', not '%d'", tx.From.String(), expectedNonce, tx.Nonce)
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("insufficient balance. Sender '%s' balance is %d TBS. Tx cost is %d TBS", tx.From.String(), s.Balances[tx.From], tx.Value)
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value
	s.AccountsToNonce[tx.From] = tx.Nonce

	return nil
}

func (s *State) copy() State {
	c := State{}
	c.Balances = make(map[common.Address]uint)
	c.AccountsToNonce = make(map[common.Address]uint)
	c.lastBlock = s.lastBlock
	c.lastBlockHash = s.lastBlockHash
	c.hasGenesisBlock = s.hasGenesisBlock

	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}

	for acc, nonce := range s.AccountsToNonce {
		c.AccountsToNonce[acc] = nonce
	}

	return c
}

func (s *State) GetNextAccountNonce(account common.Address) uint {
	return s.AccountsToNonce[account] + 1
}

func (s *State) Close() {
	s.dbFile.Close()
}
