package node

import (
	com "github.com/hyperorchid/go-miner-pool/common"
)
type Conf struct {
	DebugMode  bool
	WalletPath string
	DBPath     string
	*com.EthereumConfig
}

var SysConf = &Conf{
	EthereumConfig:com.TestNet,
}
