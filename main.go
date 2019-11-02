package main

import (
	"fmt"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner/node"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	DefaultBaseDir = ".miner"
	WalletFile     = "wallet.json"
	DataBase       = "Receipts"
)

var param struct {
	version  bool
	password string
}

var rootCmd = &cobra.Command{
	Use: "miner",

	Short: "miner",

	Long: `usage description`,

	Run: mainRun,
}

func init() {

	rootCmd.Flags().BoolVarP(&param.version, "version",
		"v", false, "show current version")

	rootCmd.Flags().BoolVarP(&node.SysConf.DebugMode, "debug",
		"d", false, "run in debug model")

	rootCmd.Flags().StringVarP(&param.password, "password",
		"p", "", "Password to open pool wallet.")

	rootCmd.AddCommand(InitCmd)
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainRun(_ *cobra.Command, _ []string) {
	base := BaseDir()
	node.SysConf.WalletPath = WalletDir(base)
	node.SysConf.DBPath = DBPath(base)

	if err := node.WInst().Open(param.password); err != nil {
		panic(err)
	}

	n := node.Inst()
	com.NewThread(n.Mining, func(err interface{}) {
		panic(err)
	}).Start()

	c := node.Chain()
	c.BM = n
	com.NewThread(c.Sync, func(err interface{}) {
		panic(err)
	}).Start()

	done := make(chan bool, 1)
	go waitSignal(done)
	<-done
}

func waitSignal(done chan bool) {
	pid := strconv.Itoa(os.Getpid())
	fmt.Printf("\n>>>>>>>>>>miner start at pid(%s)<<<<<<<<<<\n", pid)
	if err := ioutil.WriteFile(".pid", []byte(pid), 0644); err != nil {
		fmt.Print("failed to write running pid", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sigCh

	node.Inst().Stop()
	fmt.Printf("\n>>>>>>>>>>process finished(%s)<<<<<<<<<<\n", sig)

	done <- true
}
