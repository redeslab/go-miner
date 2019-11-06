package main

import (
	"fmt"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner-pool/network"
	"github.com/hyperorchid/go-miner/node"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	DefaultBaseDir = ".hop"
	WalletFile     = "wallet.json"
	DataBase       = "Receipts"
)

var param struct {
	version  bool
	password string
	minerIP  string
	basIP    string
}

var rootCmd = &cobra.Command{
	Use: "HOP",

	Short: "HOP",

	Long: `usage description`,

	Run: mainRun,
}

func init() {

	rootCmd.Flags().BoolVarP(&param.version, "version",
		"v", false, "HOP -v")

	rootCmd.Flags().BoolVarP(&node.SysConf.DebugMode, "debug",
		"d", false, "HOP -d")

	rootCmd.Flags().StringVarP(&param.password, "password",
		"p", "", "HOP -p [PASSWORD]")

	//TODO:: mv to config file
	rootCmd.Flags().StringVarP(&node.SysConf.BAS, "basIP",
		"b", "167.179.112.108", "HOP -b [BAS IP]")

	rootCmd.AddCommand(InitCmd)
	rootCmd.AddCommand(BasCmd)
	rootCmd.AddCommand(ShowCmd)
	ShowCmd.AddCommand(ShowAddrCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainRun(_ *cobra.Command, _ []string) {
	base := BaseDir()

	if _, ok := com.FileExists(base); !ok {
		fmt.Println("Init node first, please!' HOP init -p [PASSWORD]'")
		return
	}

	node.SysConf.WalletPath = WalletDir(base)
	node.SysConf.DBPath = DBPath(base)
	network.BASInst().SetServerIP(node.SysConf.BAS)
	if param.password == "" {
		fmt.Println("Password=>")

		pw, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		param.password = string(pw)
	}

	if err := node.WInst().Open(param.password); err != nil {
		panic(err)
	}

	n := node.SrvNode()
	com.NewThread(n.Mining, func(err interface{}) {
		panic(err)
	}).Start()

	c := node.Chain()
	c.BucketManager = n
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

	node.SrvNode().Stop()
	fmt.Printf("\n>>>>>>>>>>process finished(%s)<<<<<<<<<<\n", sig)

	done <- true
}
