package main

import (
	"fmt"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner/node"
	"github.com/spf13/cobra"
	"os"
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
	if _, ok := common.FileExists(baseDir); ok {
		fmt.Println("Duplicate init operation")
		return
	}
	if len(param.password) == 0 {
		pwd, err := common.ReadPassWord2()
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
}
