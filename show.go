package main

import (
	"fmt"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "show miner's basic info",
	Long:  `TODO::.`,
	//Run:   basReg,
}

var ShowAddrCmd = &cobra.Command{
	Use:   "address",
	Short: "hop miner's network layer address",
	Long:  `TODO::.`,
	Run:   showAddr,
}

func showAddr(_ *cobra.Command, _ []string) {
	w, err := account.LoadWallet(WalletDir(BaseDir()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(w.SubAddress().String())
}
