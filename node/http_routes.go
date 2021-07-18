package node

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jTanG0506/go-blockchain/database"
	"github.com/jTanG0506/go-blockchain/wallet"
)

type ErrorRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash           `json:"block_hash"`
	Balances map[common.Address]uint `json:"balances"`
}

type StatusRes struct {
	Hash       database.Hash       `json:"block_hash"`
	Number     uint64              `json:"block_number"`
	KnownPeers map[string]PeerNode `json:"peers_known"`
	PendingTXs []database.SignedTx `json:"pending_txs"`
}

type AddTXReq struct {
	From    string `json:"from"`
	FromPwd string `json:"from_pwd"`
	To      string `json:"to"`
	Value   uint   `json:"value"`
	Data    string `json:"data"`
}

type AddTXRes struct {
	Success bool `json:"success"`
}

type SyncRes struct {
	Blocks []database.Block `json:"blocks"`
}

type AddPeerRes struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func listBalancesHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalancesRes{state.LatestBlockHash(), state.Balances})
}

func statusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	res := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Number:     node.state.LastBlock().Header.Number,
		KnownPeers: node.knownPeers,
		PendingTXs: node.getPendingTXsAsArray(),
	}

	writeRes(w, res)
}

func txAddHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	req := AddTXReq{}
	err := readRequest(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	from := database.NewAccount(req.From)
	if from.String() == common.HexToAddress("").String() {
		writeRes(w, fmt.Errorf("%s is an invalid 'from' sender", from.String()))
		return
	}

	if req.FromPwd == "" {
		writeErrRes(w, fmt.Errorf("password to decrypt account %s is required, but 'from_pwd' is empty", from.String()))
		return
	}

	nonce := node.state.GetNextAccountNonce(from)
	tx := database.NewTx(
		from,
		database.NewAccount(req.To),
		nonce,
		req.Value,
		req.Data,
	)

	signedTx, err := wallet.SignTxWithKeystoreAccount(
		tx,
		from,
		req.FromPwd,
		wallet.GetKeystoreDirPath(node.dataDir),
	)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	err = node.AddPendingTX(signedTx, node.info)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, AddTXRes{Success: true})
}

func syncHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	reqHash := r.URL.Query().Get(syncEndpointQueryKeyFromBlock)
	hash := database.Hash{}

	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		writeErrRes(w, err)
		return
	}

	blocks, err := database.GetBlocksAfter(hash, node.dataDir)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, SyncRes{Blocks: blocks})
}

func addPeerHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	peerIP := r.URL.Query().Get(addPeerEndpointQueryKeyIP)
	peerPortRaw := r.URL.Query().Get(addPeerEndpointQueryKeyPort)
	minerRaw := r.URL.Query().Get(addPeerEndpointQueryKeyMiner)

	peerPort, err := strconv.ParseUint(peerPortRaw, 10, 32)
	if err != nil {
		writeRes(w, AddPeerRes{false, err.Error()})
		return
	}

	peer := NewPeerNode(peerIP, peerPort, false, database.NewAccount(minerRaw), true)
	node.AddPeer(peer)
	fmt.Printf("Peer '%s' was added into KnownPeers\n", peer.TcpAddress())

	writeRes(w, AddPeerRes{true, ""})
}
