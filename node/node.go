package node

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jTanG0506/go-blockchain/database"
)

const DefaultHTTPPort = 8080
const statusEndpoint = "/node/status"
const syncEndpoint = "/node/sync"
const syncEndpointQueryKeyFromBlock = "fromBlock"

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

type Node struct {
	dataDir    string
	port       uint64
	state      *database.State
	knownPeers map[string]PeerNode
}

func NewNode(dataDir string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: knownPeers,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Printf("Listening on HTTP port: %d\n", n.port)
	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state
	go n.sync(ctx)

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc(statusEndpoint, func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	http.HandleFunc(syncEndpoint, func(w http.ResponseWriter, r *http.Request) {
		syncHandler(w, r, n.dataDir)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}
