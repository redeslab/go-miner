package node

import (
	"github.com/ethereum/go-ethereum/common"
	com "github.com/hyperorchid/go-miner-pool/common"
	"os/user"
	"path/filepath"
)

type Conf struct {
	DebugMode  bool
	WalletPath string
	DBPath     string
	LogPath    string
	BAS        string
	PidPath    string
	*com.EthereumConfig
}

const (
	DefaultBaseDir = ".hop"
	WalletFile     = "wallet.json"
	DataBase       = "Receipts"
	LogFile        = "hop.log"
	PidFile        = "hop.pid"
)

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

func BaseDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	baseDir := filepath.Join(usr.HomeDir, string(filepath.Separator), DefaultBaseDir)
	return baseDir
}

func WalletDir(base string) string {
	return filepath.Join(base, string(filepath.Separator), WalletFile)
}

func (c *Conf) InitPath(base string) {
	c.WalletPath = filepath.Join(base, string(filepath.Separator), WalletFile)
	c.DBPath = filepath.Join(base, string(filepath.Separator), DataBase)
	c.LogPath = filepath.Join(base, string(filepath.Separator), LogFile)
	c.PidPath = filepath.Join(base, string(filepath.Separator), PidFile)
}
