package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/redeslab/go-miner-pool/account"
	"github.com/redeslab/go-miner/node"
	"github.com/redeslab/go-miner/pbs"
	"github.com/spf13/cobra"
	"math/big"
	"time"
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

var ShowCounterCmd = &cobra.Command{
	Use:   "counter",
	Short: "hop miner's network layer address",
	Long:  `TODO::.`,
	Run:   showCounter,
}

var ShowAllUser = &cobra.Command{
	Use:   "alluser",
	Short: "show all user in miner",
	Long:  `TODO::.`,
	Run:   showAlluser,
}

var ShowOneUser = &cobra.Command{
	Use:   "user",
	Short: "show user detail info",
	Long:  `TODO::.`,
	Run:   showOneuser,
}

var ShowAllReceipt = &cobra.Command{
	Use:   "allreceipt",
	Short: "show all receipt of user",
	Long:  `TODO::.`,
	Run:   showAllReceipt,
}

var ShowLatestReceipt = &cobra.Command{
	Use:   "latestreceipt",
	Short: "show latest receipt of user",
	Long:  `TODO::.`,
	Run:   showLatestReceipt,
}

var ShowOneReceipt = &cobra.Command{
	Use:   "receipt",
	Short: "show a receipt of user",
	Long:  `TODO::.`,
	Run:   showoneReceipt,
}

func init() {
	//rootCmd.AddCommand(ShowCmd)
	ShowCmd.AddCommand(ShowAddrCmd)
	ShowCmd.AddCommand(ShowCounterCmd)
	ShowCmd.AddCommand(ShowAllUser)
	ShowCmd.AddCommand(ShowOneUser)
	ShowCmd.AddCommand(ShowAllReceipt)
	ShowCmd.AddCommand(ShowLatestReceipt)
	ShowCmd.AddCommand(ShowOneReceipt)

	ShowCounterCmd.Flags().StringVarP(&param.user, "user",
		"u", "", "User's main address to show")
	ShowOneUser.Flags().StringVarP(&param.user, "user", "u", "", "User's main address")
	ShowAllReceipt.Flags().StringVarP(&param.user, "user", "u", "", "user's main address")
	ShowAllReceipt.Flags().IntVarP(&param.report, "report", "r", 0, "show report to pool tx, set 1, local tx set to 0")
	ShowLatestReceipt.Flags().StringVarP(&param.user, "user", "u", "", "user's main address")
	ShowLatestReceipt.Flags().IntVarP(&param.report, "report", "r", 0, "show report to pool tx, set 1, local tx set to 0")
	ShowOneReceipt.Flags().StringVarP(&param.user, "user", "u", "", "user's main address")
	ShowOneReceipt.Flags().IntVarP(&param.report, "report", "r", 0, "show report to pool tx, set 1, local tx set to 0")
	ShowOneReceipt.Flags().StringVarP(&param.credit, "credit", "c", "", "miner credit number of a receipt")

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

func showCounter(_ *cobra.Command, _ []string) {
	c := DialToCmdService()
	for {
		b, e := c.ShowUserCounter(context.Background(), &pbs.UserCounterReq{
			User: param.user,
		})
		if e != nil {
			fmt.Println(e)
			return
		}
		fmt.Printf("\nBucket ID:[%d] Level:[%d]", b.Id, b.Bucket)
		time.Sleep(time.Second)
	}
}

func showAlluser(_ *cobra.Command, _ []string) {
	c := DialToCmdService()
	msg, err := c.ShowAlluser(context.TODO(), &pbs.EmptyReq{})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(msg.Msg)
}

func showOneuser(_ *cobra.Command, _ []string) {
	if param.user == "" {
		fmt.Println("please enter user address")
		return
	}

	c := DialToCmdService()
	msg, err := c.ShowOneUser(context.TODO(), &pbs.UserInfoReq{
		User: param.user,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(msg.Msg)

}

func showAllReceipt(_ *cobra.Command, _ []string) {
	if param.user == "" {
		fmt.Println("please enter user address")
		return
	}

	if param.report != 0 && param.report != 1 {
		fmt.Println("report must set 0 or 1")
		return
	}

	c := DialToCmdService()
	msg, err := c.ShowAllReceipt(context.TODO(), &pbs.ReceiptReq{
		User:   param.user,
		Report: int32(param.report),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(msg.Msg)

}

func showLatestReceipt(_ *cobra.Command, _ []string) {
	if param.user == "" {
		fmt.Println("please enter user address")
		return
	}

	if param.report != 0 && param.report != 1 {
		fmt.Println("report must set 0 or 1")
		return
	}

	c := DialToCmdService()
	msg, err := c.ShowLatestReceipt(context.TODO(), &pbs.ReceiptReq{
		User:   param.user,
		Report: int32(param.report),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(msg.Msg)

}

func showoneReceipt(_ *cobra.Command, _ []string) {
	if param.user == "" {
		fmt.Println("please enter user address")
		return
	}

	if param.report != 0 && param.report != 1 {
		fmt.Println("report must set 0 or 1")
		return
	}

	if param.credit == "" {
		fmt.Println("please enter credit number")
		return
	}

	z := &big.Int{}
	if _, ok := z.SetString(param.credit, 10); !ok {
		fmt.Println("please enter correct credit number")
		return
	}

	c := DialToCmdService()
	msg, err := c.ShowOneReceipt(context.TODO(), &pbs.ReceiptOneReq{
		User:   param.user,
		Report: int32(param.report),
		Credit: param.credit,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(msg.Msg)

}

func (s *cmdService) ShowUserCounter(ctx context.Context, req *pbs.UserCounterReq) (result *pbs.CounterResult, err error) {
	b := node.SrvNode().ShowUserBucket(req.User)
	if b == nil {
		return &pbs.CounterResult{
			Id:     0,
			Bucket: 0,
		}, fmt.Errorf("no such user's bucket")
	}

	b.RLock()
	defer b.RUnlock()

	return &pbs.CounterResult{
		Id:     int32(b.BID),
		Bucket: int32(b.Token),
	}, nil
}

func (s *cmdService) SetLogLevel(ctx context.Context, req *pbs.LogLevel) (result *pbs.CommonResponse, err error) {
	return nil, nil
}

func (s *cmdService) ShowAlluser(context.Context, *pbs.EmptyReq) (*pbs.CommonResponse, error) {
	msg := node.SrvNode().UserManagement().ShowAllUser()

	return &pbs.CommonResponse{
		Msg: msg,
	}, nil
}

func (s *cmdService) ShowOneUser(c context.Context, r *pbs.UserInfoReq) (*pbs.CommonResponse, error) {

	uaddr := common.HexToAddress(r.User)

	msg := node.SrvNode().UserManagement().ShowUser(uaddr)

	return &pbs.CommonResponse{
		Msg: msg,
	}, nil

}

func (s *cmdService) ShowAllReceipt(c context.Context, r *pbs.ReceiptReq) (*pbs.CommonResponse, error) {
	uaddr := common.HexToAddress(r.User)

	msg := node.SrvNode().UserManagement().ShowAllReceipt(uaddr, int(r.Report))

	return &pbs.CommonResponse{
		Msg: msg,
	}, nil
}

func (s *cmdService) ShowLatestReceipt(c context.Context, r *pbs.ReceiptReq) (*pbs.CommonResponse, error) {
	uaddr := common.HexToAddress(r.User)

	msg := node.SrvNode().UserManagement().ShowLatestReceipt(uaddr, int(r.Report))

	return &pbs.CommonResponse{
		Msg: msg,
	}, nil
}

func (s *cmdService) ShowOneReceipt(c context.Context, r *pbs.ReceiptOneReq) (*pbs.CommonResponse, error) {
	uaddr := common.HexToAddress(r.User)
	msg := node.SrvNode().UserManagement().ShowReceipt(uaddr, r.Credit, int(r.Report))

	return &pbs.CommonResponse{
		Msg: msg,
	}, nil

}
