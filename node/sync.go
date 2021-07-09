package node

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(45 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Searching for new peers and blocks...")
			n.fetchNewBlocksAndPeers()
		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) fetchNewBlocksAndPeers() {
	for _, peer := range n.knownPeers {
		status, err := queryPeerStatus(peer)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}

		localBlockNumber := n.state.LastBlock().Header.Number
		if localBlockNumber < status.Number {
			newBlocksCount := status.Number - localBlockNumber
			fmt.Printf("Found %d new blocks from peer %s\n", newBlocksCount, peer.IP)
		}

		for _, possibleNewPeer := range status.KnownPeers {
			_, isKnownPeer := n.knownPeers[possibleNewPeer.TcpAddress()]
			if !isKnownPeer {
				fmt.Printf("Found a new peer %s\n", possibleNewPeer.TcpAddress())
				n.knownPeers[possibleNewPeer.TcpAddress()] = possibleNewPeer
			}
		}
	}
}

func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	url := fmt.Sprintf("http://%s/%s", peer.TcpAddress(), statusEndpoint)
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, err
	}

	statusRes := StatusRes{}
	err = readResponse(res, &statusRes)
	if err != nil {
		return StatusRes{}, err
	}

	return statusRes, nil
}
