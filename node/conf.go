package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	com "github.com/redeslab/go-miner-pool/common"
	"github.com/redeslab/go-miner/util"
	"github.com/redeslab/pirate_contract/config"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
)

type PathConf struct {
	WalletPath string
	DBPath     string
	LogPath    string
	PidPath    string
	ConfPath   string

	WebAuthPath       string
	WebAuthTokenPath  string
	WebAuthVerifyPath string

	WebMinerPath    string
	WebMinerDetails string
	WebUserPath     string
	WebUserCount    string
	WebUserInfo     string
}

var HopVersion string = "1.1.6"

var accessPubKeyLock sync.Mutex

type SettingConf struct {
	BAS string
	*com.EthereumConfig
	WebPort      int
	AccessPubKey []string
}

type MinerConf struct {
	BAS          string `json:"bas"`
	ECfg         map[int]*com.EthereumConfig
	WebPort      int      `json:"web_port"`
	AccessPubKey []string `json:"access_pub_key"`
}

func (mc *MinerConf) String() string {
	j, _ := json.MarshalIndent(*mc, " ", " \t")
	return string(j)
}

//type MinerConf struct {
//	WebPort      int   `json:"web_port,omitempty"`
//	AccessPubKey []string `json:"access_addr,omitempty"`
//}

const (
	DefaultBaseDir = ".hop"
	WalletFile     = "wallet.json"
	DataBase       = "Receipts"
	LogFile        = "log.hop"
	PidFile        = "pid.hop"
	ConfFile       = "conf.hop"
	WebPort        = 42887
	LMCMD          = "cmd"
)

var CMDServicePort = "42776"
var MinerSetting = &SettingConf{EthereumConfig: &com.EthereumConfig{}}
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

func MinerConfFile(bas string) string {
	return filepath.Join(bas, string(filepath.Separator), ConfFile)
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

	pc.WebAuthPath = "/auth"
	pc.WebAuthTokenPath = "/token"
	pc.WebAuthVerifyPath = "/verify"

	pc.WebMinerPath = "/miner"
	pc.WebMinerDetails = "/info"
	pc.WebUserPath = "/user"
	pc.WebUserCount = "/count"
	pc.WebUserInfo = "/info"

	fmt.Println(pc.String())
}

func InitEthConfig() {
	if MinerSetting.EthereumConfig == nil {
		panic("init sys setting first")
	}
	config.InitEthConf(MinerSetting.EthereumConfig.NetworkID,
		MinerSetting.EthereumConfig.EthApiUrl,
		MinerSetting.EthereumConfig.Token.String(),
		MinerSetting.EthereumConfig.MicroPaySys.String(),
		nil)
}

func InitMinerNode(auth, port string, networkid int) {
	PathSetting.InitPath()

	jsonStr, err := ioutil.ReadFile(PathSetting.ConfPath)
	if err != nil {
		panic("Load config failed")
	}

	mc := &MinerConf{}

	if err := json.Unmarshal(jsonStr, mc); err != nil {
		panic(err)
	}

	if networkid == com.MainNetworkId {
		MinerSetting.NetworkID = com.MainNetworkId
		MinerSetting.Token = mc.ECfg[com.MainNetworkId].Token
		MinerSetting.EthApiUrl = mc.ECfg[com.MainNetworkId].EthApiUrl
		MinerSetting.MicroPaySys = mc.ECfg[com.MainNetworkId].MicroPaySys
	} else {
		MinerSetting.NetworkID = com.RopstenNetworkId
		MinerSetting.Token = mc.ECfg[com.RopstenNetworkId].Token
		MinerSetting.EthApiUrl = mc.ECfg[com.RopstenNetworkId].EthApiUrl
		MinerSetting.MicroPaySys = mc.ECfg[com.RopstenNetworkId].MicroPaySys
	}

	MinerSetting.BAS = mc.BAS
	MinerSetting.WebPort = mc.WebPort
	MinerSetting.AccessPubKey = mc.AccessPubKey

	fmt.Println(mc.String())
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

func (cf *SettingConf) Save() error {
	confData, err := ioutil.ReadFile(PathSetting.ConfPath)
	if err != nil {
		return err
	}

	conf := &MinerConf{}
	if err := json.Unmarshal(confData, conf); err != nil {
		return err
	}

	conf.AccessPubKey = cf.AccessPubKey
	conf.WebPort = cf.WebPort
	j, _ := json.MarshalIndent(conf, " ", "\t")

	return util.Save2File(j, PathSetting.ConfPath)
}

func (cf *SettingConf) GetAccessAddrs2() []string {
	accessPubKeyLock.Lock()
	defer accessPubKeyLock.Unlock()

	return cf.AccessPubKey
}

func (cf *SettingConf) SetWebPort(webPort int) {
	cf.WebPort = webPort
}

func (cf *SettingConf) GetWebPort() int {
	return cf.WebPort
}

func (cf *SettingConf) AddAccessAddr(addr string) error {
	accessPubKeyLock.Lock()
	defer accessPubKeyLock.Unlock()

	if !common.IsHexAddress(addr) {
		return errors.New("access address error")
	}

	cf.AccessPubKey = append(cf.AccessPubKey, addr)

	return nil
}

func (cf *SettingConf) RemoveAccessAddr(addr string) error {
	accessPubKeyLock.Lock()
	defer accessPubKeyLock.Unlock()

	if !common.IsHexAddress(addr) {
		return errors.New("access address error")
	}

	idx := -1
	for i := 0; i < len(cf.AccessPubKey); i++ {
		if cf.AccessPubKey[i] == addr {
			idx = i
		}
	}
	if idx == -1 {
		return errors.New("address not exists")
	}

	l := len(cf.AccessPubKey) - 1
	if idx != l {
		cf.AccessPubKey[idx] = cf.AccessPubKey[l]
	}

	cf.AccessPubKey = cf.AccessPubKey[:l]

	return nil
}

func (ss *SettingConf) GetAccessAddrs() string {
	accessPubKeyLock.Lock()
	defer accessPubKeyLock.Unlock()

	msg := ""

	for i := 0; i < len(ss.AccessPubKey); i++ {
		if msg != "" {
			msg += "\r\n"
		}
		msg += strconv.Itoa(i + 1)
		msg += " : "
		msg += ss.AccessPubKey[i]
	}

	return msg
}
