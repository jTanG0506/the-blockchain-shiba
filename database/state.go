package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile          *os.File
	lastBlock       Block
	lastBlockHash   Hash
	hasGenesisBlock bool
}

func NewStateFromDisk(dataDir string) (*State, error) {
	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	blocksFilePath := getBlocksDbFilePath(dataDir)
	blocks, err := os.OpenFile(blocksFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(blocks)
	state := &State{balances, make([]Tx, 0), blocks, Block{}, Hash{}, false}

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

		err = applyTXs(blockFs.Value.TXs, state)
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

	err := applyBlock(b, tempState)
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
	s.lastBlockHash = tempState.lastBlockHash
	s.lastBlock = tempState.lastBlock
	s.hasGenesisBlock = true

	return blockHash, nil
}

func (s *State) AddTx(tx Tx) error {
	if err := applyTx(tx, s); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)
	return nil
}

func applyBlock(b Block, s State) error {
	expectedNextBlockNumber := s.lastBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != expectedNextBlockNumber {
		return fmt.Errorf("next expected block must have number '%d' not '%d'", expectedNextBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.lastBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.lastBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.lastBlockHash, b.Header.Parent)
	}

	return applyTXs(b.TXs, &s)
}

func applyTXs(txs []Tx, s *State) error {
	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyTx(tx Tx, s *State) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("insufficient balance. Sender '%s' balance is %d TBS. Tx cost is %d TBS", tx.From, s.Balances[tx.From], tx.Value)
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

func (s *State) copy() State {
	c := State{}
	c.Balances = make(map[Account]uint)
	c.txMempool = make([]Tx, len(s.txMempool))
	c.lastBlock = s.lastBlock
	c.lastBlockHash = s.lastBlockHash
	c.hasGenesisBlock = s.hasGenesisBlock

	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}

	for _, tx := range s.txMempool {
		c.txMempool = append(c.txMempool, tx)
	}

	return c
}

func (s *State) Close() {
	s.dbFile.Close()
}
