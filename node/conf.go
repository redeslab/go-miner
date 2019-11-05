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
	EthereumConfig: &com.EthereumConfig{
		NetworkID:   com.RopstenNetworkId,
		EthApiUrl:   "https://ropsten.infura.io/v3/f3245cef90ed440897e43efc6b3dd0f7",
		MicroPaySys: "0x95048537a137ac8bf4d612824e0f74fbae34542a",
		Token:       "0x19e55b69597e4ad2e99a422af70206cc453c3561",
	},
	BAS: "149.28.203.172",
}
