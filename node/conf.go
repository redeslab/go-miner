package node

import (
	com "github.com/hyperorchid/go-miner-pool/common"
)

type Conf struct {
	DebugMode  bool
	WalletPath string
	DBPath     string
	BAS        string
	*com.EthereumConfig
}

var SysConf = &Conf{
	EthereumConfig: com.TestNet,
	BAS:            "8.8.8.8",
}
