package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redeslab/go-miner-pool/account"
	com "github.com/redeslab/go-miner-pool/common"
	"github.com/redeslab/go-miner/node"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "init miner node",
	Long:  `TODO::.`,
	Run:   initMiner,
}

func init() {
	InitCmd.Flags().StringVarP(&param.password, "password", "p", "", "Password to create Hyper Orchid block chain system.")
}
func initMiner(_ *cobra.Command, _ []string) {

	baseDir := node.BaseDir()
	if _, ok := com.FileExists(baseDir); ok {
		fmt.Println("Duplicate init operation")
		return
	}
	if len(param.password) == 0 {
		pwd, err := com.ReadPassWord2()
		if err != nil {
			panic(err)
		}
		param.password = pwd
	}

	if err := os.Mkdir(baseDir, os.ModePerm); err != nil {
		panic(err)
	}

	w, err := account.NewWallet(param.password)
	if err != nil {
		panic(err)
	}

	if err := w.SaveToPath(node.WalletDir(baseDir)); err != nil {
		panic(err)
	}
	fmt.Println("Create wallet success!")

	defaultSys := &node.MinerConf{
		BAS:     "167.179.75.39",
		WebPort: node.WebPort,
		ECfg:    make(map[int]*com.EthereumConfig),
	}

	defaultSys.ECfg[com.RopstenNetworkId] = &com.EthereumConfig{
		NetworkID:   com.RopstenNetworkId,
		EthApiUrl:   "https://ropsten.infura.io/v3/d64d364124684359ace20feae1f9ac20",
		MicroPaySys: common.HexToAddress("0x72d5f9f633f537f87ef7415b8bdbfa438d0a1a6c"),
		Token:       common.HexToAddress("0xad44c8493de3fe2b070f33927a315b50da9a0e25"),
	}

	defaultSys.ECfg[com.MainNetworkId] = &com.EthereumConfig{
		NetworkID:   com.MainNetworkId,
		EthApiUrl:   "https://mainnet.infura.io/v3/d64d364124684359ace20feae1f9ac20",
		MicroPaySys: common.HexToAddress("0x5eC8589C8832AdE0B45C608681bbaCEF517E7caD"),
		Token:       common.HexToAddress("0x1999ac2b141e6d5c4e27579b30f842078bc620b3"),
	}

	byt, err := json.MarshalIndent(defaultSys, "", "\t")
	confPath := filepath.Join(baseDir, string(filepath.Separator), node.ConfFile)
	if err := ioutil.WriteFile(confPath, byt, 0644); err != nil {
		panic(err)
	}
}
