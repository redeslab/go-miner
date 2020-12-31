package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hyperorchidlab/go-miner-pool/account"
	"github.com/hyperorchidlab/go-miner/node"
	"github.com/hyperorchidlab/go-miner/pbs"
	"github.com/spf13/cobra"
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

func init() {
	//rootCmd.AddCommand(ShowCmd)
	ShowCmd.AddCommand(ShowAddrCmd)
	ShowCmd.AddCommand(ShowCounterCmd)

	ShowCounterCmd.Flags().StringVarP(&param.user, "user",
		"u", "", "User's main address to show")

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
