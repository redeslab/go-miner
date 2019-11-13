package node

import (
	"github.com/ethereum/go-ethereum/common"
	com "github.com/hyperorchid/go-miner-pool/common"
)

type Conf struct {
	DebugMode  bool
	WalletPath string
	DBPath     string
	LogPath    string
	BAS        string
	*com.EthereumConfig
}

//TODO::
var SysConf = &Conf{
	EthereumConfig: &com.EthereumConfig{
		NetworkID:   com.RopstenNetworkId,
		EthApiUrl:   "https://ropsten.infura.io/v3/f3245cef90ed440897e43efc6b3dd0f7",
		MicroPaySys: common.HexToAddress("0x9a04dC6d9DE10F6404CaAfbe3F80e70f2dAec7DB"),
		Token:       common.HexToAddress("0x3adc98d5e292355e59ae2ca169d241d889b092e3"),
	},
	BAS: "167.179.112.108",
}
