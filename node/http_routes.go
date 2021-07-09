package node

import (
	"net/http"

	"github.com/jTanG0506/go-blockchain/database"
)

type ErrorRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type StatusRes struct {
	Hash       database.Hash       `json:"block_hash"`
	Number     uint64              `json:"block_number"`
	KnownPeers map[string]PeerNode `json:"peers_known"`
}

type AddTXReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type AddTXRes struct {
	Hash database.Hash `json:"block_hash"`
}

func listBalancesHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalancesRes{state.LatestBlockHash(), state.Balances})
}

func statusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	res := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Number:     node.state.LastBlock().Header.Number,
		KnownPeers: node.knownPeers,
	}

	writeRes(w, res)
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	req := AddTXReq{}
	err := readRequest(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	tx := database.NewTx(
		database.NewAccount(req.From),
		database.NewAccount(req.To),
		req.Value,
		req.Data,
	)
	err = state.AddTx(tx)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	hash, err := state.Persist()
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, AddTXRes{hash})
}
