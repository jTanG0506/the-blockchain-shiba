package node

import (
	"fmt"
	"net/http"

	"github.com/jTanG0506/go-blockchain/database"
)

const DefaultHTTPPort = 8080

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

type Node struct {
	dataDir    string
	port       uint64
	state      *database.State
	knownPeers []PeerNode
}

func NewNode(dataDir string, port uint64, bootstrap PeerNode) *Node {
	return &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: []PeerNode{bootstrap},
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

func (n *Node) Run() error {
	fmt.Printf("Listening on HTTP port: %d\n", n.port)
	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	n.state = state

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/node/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}
