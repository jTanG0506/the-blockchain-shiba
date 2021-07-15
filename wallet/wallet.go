package wallet

import "path/filepath"

const keystoreDirName = "keystore"

const ToshiAccount = "0xe5ED8C1829192380205b1E7BB5A3F44baf181d25"
const JTangAccount = "0xf70D226203FDDa745C3B160D92Ee665A71191D6a"
const QudsiiAccount = "0x7573428c0394133cC5A3FC5533b9B04241D1271E"

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}
