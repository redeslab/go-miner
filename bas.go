package main

import (
	"encoding/json"
	"fmt"
	basc "github.com/hyperorchidlab/BAS/client"
	"github.com/hyperorchidlab/BAS/crypto"
	"github.com/hyperorchidlab/BAS/dbSrv"
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

	BasCmd.Flags().StringVarP(&param.location, "location","l","","set miner location")

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

	if param.location == "" || len(param.location)>8{
		fmt.Println("please set miner location, and not more than 8 bytes")
		return
	}

	extData := &bas.MinerExtendData{}
	extData.HopAddr = node.WInst().SubAddress().String()
	extData.MainAddr = node.WInst().MainAddress().String()
	extData.Location = param.location

	basip:=param.basIP

	if basip == ""{
		node.PathSetting.ConfPath = node.MinerConfFile(node.BaseDir())
		jsonStr, err := ioutil.ReadFile(node.PathSetting.ConfPath)
		if err != nil {
			panic("Load config failed")
		}
		if err := json.Unmarshal(jsonStr, node.SysConf); err != nil {
			panic(err)
		}

		basip = node.SysConf.BAS
		if net.ParseIP(basip) == nil{
			panic("bas ip from config file error")
		}

	}

	req := &dbSrv.RegRequest{
		BlockAddr: []byte(extData.HopAddr),
		SignData:dbSrv.SignData{
			NetworkAddr:&dbSrv.NetworkAddr{
				NTyp:    t,
				NetAddr: []byte(param.minerIP),
				BTyp:    crypto.HOP,
			},
			ExtData:extData.Marshal(),
		},
	}

	req.Sig = node.WInst().SignJSONSub(req.SignData)
	if err := basc.RegisterBySrvIP(req, basip); err != nil {
		panic(err)
	}
	fmt.Println("reg success!")
}
