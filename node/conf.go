package node

type Conf struct {
	DebugMode  bool
	WalletPath string
	DBPath     string
}

var SysConf = &Conf{}
