package main

import (
	"fmt"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner-pool/common"
	"github.com/spf13/cobra"
	"os"
	"os/user"
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

	baseDir := BaseDir()
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

	if err := w.SaveToPath(WalletDir(baseDir)); err != nil {
		panic(err)
	}
	fmt.Println("Create wallet success!")
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
func DBPath(base string) string {
	return filepath.Join(base, string(filepath.Separator), DataBase)
}
