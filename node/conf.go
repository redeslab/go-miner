package node

import (
	"encoding/json"
	"fmt"
	com "github.com/hyperorchid/go-miner-pool/common"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type PathConf struct {
	WalletPath string
	DBPath     string
	LogPath    string
	PidPath    string
	ConfPath   string
}

type Conf struct {
	BAS string
	*com.EthereumConfig
}

const (
	DefaultBaseDir = ".hop"
	WalletFile     = "wallet.json"
	DataBase       = "Receipts"
	LogFile        = "log.hop"
	PidFile        = "pid.hop"
	ConfFile       = "conf.hop"
)

var CMDServicePort = "42017"
var SysConf = &Conf{}
var PathSetting = &PathConf{}

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

func (pc *PathConf) String() string {
	return fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++\n"+
		"+WalletPath:\t%s+\n"+
		"+DBPath:\t%s+\n"+
		"+LogPath:\t%s+\n"+
		"+PidPath:\t%s+\n"+
		"+ConfPath:\t%s+\n"+
		"++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
		pc.WalletPath,
		pc.DBPath,
		pc.LogPath,
		pc.PidPath,
		pc.ConfPath)
}

func (pc *PathConf) InitPath() {
	base := BaseDir()
	if _, ok := com.FileExists(base); !ok {
		panic("Init node first, please!' HOP init -p [PASSWORD]'")
	}
	pc.WalletPath = filepath.Join(base, string(filepath.Separator), WalletFile)
	pc.DBPath = filepath.Join(base, string(filepath.Separator), DataBase)
	pc.LogPath = filepath.Join(base, string(filepath.Separator), LogFile)
	pc.PidPath = filepath.Join(base, string(filepath.Separator), PidFile)
	pc.ConfPath = filepath.Join(base, string(filepath.Separator), ConfFile)
	fmt.Println(pc.String())
}

func InitMinerNode(auth, port string) {
	PathSetting.InitPath()

	jsonStr, err := ioutil.ReadFile(PathSetting.ConfPath)
	if err != nil {
		panic("Load config failed")
	}
	if err := json.Unmarshal(jsonStr, SysConf); err != nil {
		panic(err)
	}

	fmt.Println(SysConf.String())
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

	com.InitLog(PathSetting.LogPath)
	CMDServicePort = port
}
