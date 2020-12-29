package main

import (
	"encoding/json"
	"fmt"
	basc "github.com/hyperorchidlab/BAS/client"
	"github.com/hyperorchidlab/BAS/crypto"
	"github.com/hyperorchidlab/BAS/dbSrv"
	"github.com/hyperorchidlab/go-miner-pool/util/privateip"
	"github.com/hyperorchidlab/go-miner/bas"
	"github.com/hyperorchidlab/go-miner/node"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
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
		"b", "", "HOP bas -b [BAS IP]]")

	BasCmd.Flags().StringVarP(&param.location, "location", "l", "", "set miner location")

}

func basReg(_ *cobra.Command, _ []string) {

	node.PathSetting.WalletPath = node.WalletDir(node.BaseDir())

	if err := node.WInst().Open(param.password); err != nil {
		fmt.Println("password not correct, can't open wallet")
		return
	}

	t, e := dbSrv.CheckIPType(param.minerIP)
	if e != nil {
		fmt.Println("ip error, ", e.Error())
		return
	}

	if privateip.IsPrivateIPStr(param.minerIP) {
		fmt.Println("error: miner ip is a reserved ip")
		return
	}

	if param.location == "" || len(param.location) > 8 {
		fmt.Println("please set miner location, and not more than 8 bytes")
		return
	}

	extData := &bas.MinerExtendData{}
	extData.HopAddr = node.WInst().SubAddress().String()
	extData.MainAddr = node.WInst().MainAddress().String()
	extData.Location = param.location

	basip := param.basIP

	if basip == "" {
		fmt.Println("bas ip not set, use system config ip address")
		node.PathSetting.ConfPath = node.MinerConfFile(node.BaseDir())
		jsonStr, err := ioutil.ReadFile(node.PathSetting.ConfPath)
		if err != nil {
			fmt.Println("load config failed")
			return
		}
		if err := json.Unmarshal(jsonStr, node.SysConf); err != nil {
			fmt.Println(err)
			return
		}

		basip = node.SysConf.BAS
		if net.ParseIP(basip) == nil {
			fmt.Println("bas ip from config file error")
			return
		}
	}

	req := &dbSrv.RegRequest{
		BlockAddr: []byte(extData.HopAddr),
		SignData: dbSrv.SignData{
			NetworkAddr: &dbSrv.NetworkAddr{
				NTyp:    t,
				NetAddr: []byte(param.minerIP),
				BTyp:    crypto.HOP,
			},
			ExtData: extData.Marshal(),
		},
	}

	req.Sig = node.WInst().SignJSONSub(req.SignData)
	if err := basc.RegisterBySrvIP(req, basip); err != nil {
		panic(err)
	}
	fmt.Println("reg success!")
}
