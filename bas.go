package main

import (
	"fmt"
	"github.com/hyperorchid/go-miner-pool/network"
	"github.com/hyperorchid/go-miner/node"
	"github.com/hyperorchidlab/BAS/dbSrv"
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
		"b", "149.28.203.172", "HOP bas -b [BAS IP]]")
}

func basReg(_ *cobra.Command, _ []string) {

	node.SysConf.WalletPath = WalletDir(BaseDir())

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
			BTyp:    dbSrv.BTEd25519,
		},
	}

	req.Sig = node.WInst().SignJSONSub(req.NetworkAddr)
	if err := network.BASInst().RegisterWithSrv(req, param.basIP); err != nil {
		panic(err)
	}
	fmt.Println("reg success!")
}
