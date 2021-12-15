package main

import (
	"context"
	"fmt"
	"github.com/op/go-logging"
	common "github.com/redeslab/go-miner/node"
	"github.com/redeslab/go-miner/pbs"
	"github.com/redeslab/go-miner/webserver"
	"github.com/spf13/cobra"
	_ "google.golang.org/grpc"
	"strconv"
	"time"
)

var webCmdParam = struct {
	webPort    int
	addAddr    string
	removeAddr string
}{}

var WebAccessAddrCmd = &cobra.Command{
	Use:   "web-access-addr",
	Short: "show web access address",
	Long:  `show web access address`,
	Run:   showWebAccessAddrs,
	//Args:  cobra.MinimumNArgs(1),
}

var WebAccessAddAddrCmd = &cobra.Command{
	Use:   "add",
	Short: "add web access address",
	Long:  `add web access address`,
	Run:   addWebAccessAddrs,
}

var WebAccessRemoveAddrCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove web access address",
	Long:  `remove web access address`,
	Run:   removeWebAccessAddrs,
}

var WebPortShowCmd = &cobra.Command{
	Use:   "port",
	Short: "show web listen port",
	Long:  `show web listen port`,
	Run:   showWebPort,
}

var WebPortChangeCmd = &cobra.Command{
	Use:   "change-port",
	Short: "change web listen port",
	Long:  `change web listen port`,
	Run:   changeWebPort,
}

var logger, _ = logging.GetLogger(common.LMCMD)

func init() {
	WebAccessAddAddrCmd.Flags().StringVarP(&webCmdParam.addAddr, "address", "a", "", "ethereum address")
	WebAccessRemoveAddrCmd.Flags().StringVarP(&webCmdParam.removeAddr, "address", "a", "", "ethereum address")
	WebPortChangeCmd.Flags().IntVarP(&webCmdParam.webPort, "port", "p", 42888, "set web port [80:65535)")
	WebAccessAddrCmd.AddCommand(WebAccessAddAddrCmd)
	WebAccessAddrCmd.AddCommand(WebAccessRemoveAddrCmd)
	WebAccessAddrCmd.AddCommand(WebPortShowCmd)
	WebAccessAddrCmd.AddCommand(WebPortChangeCmd)
}

func showWebAccessAddrs(_ *cobra.Command, _ []string) {
	c := DialToCmdService()
	ua, e := c.ShowAccessAddr(context.Background(), &pbs.EmptyRequest{})
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ua.Msg)
}

func addWebAccessAddrs(_ *cobra.Command, _ []string) {
	if webCmdParam.addAddr == "" {
		fmt.Println("ethereum address is not set")
		return
	}

	c := DialToCmdService()
	resp, e := c.AccessAddressMgmt(context.Background(), &pbs.AccessAddress{Adddr: webCmdParam.addAddr, Op: 1})
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(resp.Msg)

}

func removeWebAccessAddrs(_ *cobra.Command, _ []string) {
	if webCmdParam.removeAddr == "" {
		fmt.Println("ethereum address is not set")
		return
	}

	c := DialToCmdService()
	resp, e := c.AccessAddressMgmt(context.Background(), &pbs.AccessAddress{Adddr: webCmdParam.removeAddr, Op: 2})
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(resp.Msg)
}

func showWebPort(_ *cobra.Command, _ []string) {
	c := DialToCmdService()
	ua, e := c.ShowWebPort(context.Background(), &pbs.EmptyRequest{})
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ua.Msg)
}

func changeWebPort(_ *cobra.Command, _ []string) {
	c := DialToCmdService()
	ua, e := c.WebPortSet(context.Background(), &pbs.WebPort{Port: int32(webCmdParam.webPort)})
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ua.Msg)
}

func (s *cmdService) ShowAccessAddr(ctx context.Context, request *pbs.EmptyRequest) (*pbs.CommonResponse, error) {
	return &pbs.CommonResponse{Msg: common.MinerSetting.GetAccessAddrs()}, nil
}

func (s *cmdService) ShowWebPort(ctx context.Context, request *pbs.EmptyRequest) (*pbs.CommonResponse, error) {
	return &pbs.CommonResponse{Msg: "web listen port: " + strconv.Itoa(common.MinerSetting.GetWebPort())}, nil
}

func (s *cmdService) WebPortSet(ctx context.Context, req *pbs.WebPort) (*pbs.CommonResponse, error) {
	if req.Port < 80 || req.Port > 65535 {
		return &pbs.CommonResponse{Msg: "port error, port range: [80:65535)"}, nil
	}

	if common.MinerSetting.GetWebPort() == int(req.Port) {
		return &pbs.CommonResponse{Msg: "setting port is same to current system web port"}, nil
	}

	common.MinerSetting.SetWebPort(int(req.Port))

	e := common.MinerSetting.Save()

	if e == nil {
		webserver.StopWebDaemon()
		time.Sleep(time.Second * 5)
		go webserver.StartWebDaemon()
		return &pbs.CommonResponse{Msg: "set success"}, nil
	}

	return &pbs.CommonResponse{Msg: "set failed" + e.Error()}, nil
}

func (s *cmdService) AccessAddressMgmt(ctx context.Context, req *pbs.AccessAddress) (*pbs.CommonResponse, error) {

	msg := "command line error"

	if req.Op == 1 {
		err := common.MinerSetting.AddAccessAddr(req.Adddr)
		if err != nil {
			msg = err.Error()
		} else {
			msg = "add access address " + req.Adddr + " success"
			e := common.MinerSetting.Save()
			if e != nil {
				logger.Fatal("save setting error " + e.Error())
			}
		}
	} else if req.Op == 2 {
		err := common.MinerSetting.RemoveAccessAddr(req.Adddr)
		if err != nil {
			msg = err.Error()
		} else {
			msg = "remove access address " + req.Adddr + " success"
			e := common.MinerSetting.Save()
			if e != nil {
				logger.Fatal("save setting error " + e.Error())
			}
		}
	}

	return &pbs.CommonResponse{Msg: msg}, nil
}
