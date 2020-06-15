package main

import (
	"fmt"
	basc "github.com/hyperorchidlab/BAS/client"
	"github.com/hyperorchidlab/BAS/crypto"
	"github.com/hyperorchidlab/BAS/dbSrv"
	"github.com/hyperorchidlab/go-miner/node"
	"github.com/spf13/cobra"
)

var BasCmd = &cobra.Command{
	Use:   "bas",
	Short: "register self to block chain service",
	Long:  `TODO::.`,
	Run:   basReg,
}

func init() {
	BasCmd.Flags().StringVarP(&param.minerIP, "minerIP",
		"m", "", "HOP bas -m [MY IP Address]")

	BasCmd.Flags().StringVarP(&param.password, "password",
		"p", "", "HOP bas -p [PASSWORD]")

	BasCmd.Flags().StringVarP(&param.basIP, "basIP",
		"b", "108.61.223.99", "HOP bas -b [BAS IP]]")
}

func basReg(_ *cobra.Command, _ []string) {

	node.PathSetting.WalletPath = node.WalletDir(node.BaseDir())

	if err := node.WInst().Open(param.password); err != nil {
		panic(err)
	}

	t, e := dbSrv.CheckIPType(param.minerIP)
	if e != nil {
		panic(e)
	}

	myAddr := node.WInst().SubAddress()
	fmt.Println(myAddr, len(myAddr))
	req := &dbSrv.RegRequest{
		BlockAddr: []byte(myAddr),
		NetworkAddr: &dbSrv.NetworkAddr{
			NTyp:    t,
			NetAddr: []byte(param.minerIP),
			BTyp:    crypto.HOP,
		},
	}

	req.Sig = node.WInst().SignJSONSub(req.NetworkAddr)
	if err := basc.RegisterBySrvIP(req, param.basIP); err != nil {
		panic(err)
	}
	fmt.Println("reg success!")
}
