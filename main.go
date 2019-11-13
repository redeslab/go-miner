package main

import (
	"fmt"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner/node"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
		"v", false, "HOP version")

	rootCmd.Flags().BoolVarP(&node.SysConf.DebugMode, "debug",
		"d", false, "Debug Mode")

	rootCmd.Flags().StringVarP(&param.password, "password",
		"p", "", "Password to unlock miner")

	//TODO:: mv to config file
	rootCmd.Flags().StringVarP(&node.SysConf.BAS, "basIP",
		"b", "167.179.112.108", "Bas IP")

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
	base := node.BaseDir()
	if _, ok := com.FileExists(base); !ok {
		fmt.Println("Init node first, please!' HOP init -p [PASSWORD]'")
		return
	}

	node.SysConf.InitPath(base)

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

	com.InitLog(node.SysConf.LogPath)

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
	if err := ioutil.WriteFile(node.SysConf.PidPath, []byte(pid), 0644); err != nil {
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
