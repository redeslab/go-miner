package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner/node"
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
	w, err := account.LoadWallet(node.WalletDir(node.BaseDir()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(w.MainAddress().String())
	fmt.Println(w.SubAddress().String())
	fmt.Println(hexutil.Encode(w.SubAddress().ToPubKey()))

}
