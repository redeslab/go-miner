package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	com "github.com/hyperorchid/go-miner-pool/common"
	"golang.org/x/crypto/ssh/terminal"
	"os"
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

var CMDServicePort = "42017"

//TODO::
var SysConf = &Conf{
	EthereumConfig: &com.EthereumConfig{
		NetworkID:   com.RopstenNetworkId,
		EthApiUrl:   "https://ropsten.infura.io/v3/f3245cef90ed440897e43efc6b3dd0f7",
		MicroPaySys: common.HexToAddress("0x4291d9Ff189D90Ba875E0fc1Da4D602406DD7D6e"),
		Token:       common.HexToAddress("0xAd44c8493dE3FE2B070f33927A315b50Da9a0e25"),
	},
	BAS: "108.61.223.99",
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

func InitMinerNode(auth, port string) {

	base := BaseDir()
	if _, ok := com.FileExists(base); !ok {
		panic("Init node first, please!' HOP init -p [PASSWORD]'")
		return
	}
	SysConf.InitPath(base)

	if auth == "" {
		fmt.Println("Password=>")
		pw, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		auth = string(pw)
	}

	if err := WInst().Open(auth); err != nil {
		panic(err)
	}

	com.InitLog(SysConf.LogPath)
	CMDServicePort = port
}
