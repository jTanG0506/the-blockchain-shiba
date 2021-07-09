package database

import (
	"bufio"
	"encoding/json"
	"os"
)

func GetBlocksAfter(hash Hash, dataDir string) ([]Block, error) {
	f, err := os.OpenFile(getBlocksDbFilePath(dataDir), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)
	shouldCollect := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blocksFs BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blocksFs)
		if err != nil {
			return nil, err
		}

		if shouldCollect {
			blocks = append(blocks, blocksFs.Value)
			continue
		}

		if hash == blocksFs.Key {
			shouldCollect = true
		}
	}

	return blocks, nil
}
