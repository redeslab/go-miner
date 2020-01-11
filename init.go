package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchid/go-miner-pool/account"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner/node"
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

	defaultSys := &node.Conf{
		EthereumConfig: &com.EthereumConfig{
			NetworkID:   com.RopstenNetworkId,
			EthApiUrl:   "https://ropsten.infura.io/v3/f3245cef90ed440897e43efc6b3dd0f7",
			MicroPaySys: common.HexToAddress("0xbabababababababababababababababababababa"),
			Token:       common.HexToAddress("0xbabababababababababababababababababababa"),
		},
		BAS: "108.61.223.99",
	}

	byt, err := json.MarshalIndent(defaultSys, "", "\t")
	confPath := filepath.Join(baseDir, string(filepath.Separator), node.ConfFile)
	if err := ioutil.WriteFile(confPath, byt, 0644); err != nil {
		panic(err)
	}
}
