package main

import (
	"fmt"
	com "github.com/redeslab/go-miner-pool/common"
	"github.com/redeslab/go-miner/node"
	"github.com/redeslab/go-miner/pbs"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var hopVersion string = "0.2"

var param struct {
	version  bool
	CMDPort  string
	password string
	minerIP  string
	basIP    string
	user     string
	location string
}

var rootCmd = &cobra.Command{
	Use:   "HOP",
	Short: "HOP",
	Long:  `usage description`,
	Run:   mainRun,
}

func init() {

	rootCmd.Flags().BoolVarP(&param.version, "version",
		"v", false, "HOP version")

	rootCmd.Flags().StringVarP(&param.password, "password",
		"p", "", "Password to unlock miner")

	rootCmd.Flags().StringVarP(&param.CMDPort, "cmdPort",
		"c", "42017", "Cmd service port")

	//TODO:: mv to config file
	rootCmd.Flags().StringVarP(&node.SysConf.BAS, "basIP",
		"b", "108.61.223.99", "Bas IP")

	rootCmd.AddCommand(InitCmd)
	rootCmd.AddCommand(BasCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainRun(_ *cobra.Command, _ []string) {

	if param.version {
		fmt.Println("Hop version: ", hopVersion)
		return
	}

	node.InitMinerNode(param.password, param.CMDPort)

	n := node.SrvNode()
	com.NewThreadWithID("[TCP Service Thread]", n.Mining, func(err interface{}) {
		panic(err)
	}).Start()

	c := node.Chain()
	c.BucketManager = n
	com.NewThreadWithID("[Micro Chain Sync Thread]", c.Sync, func(err interface{}) {
		panic(err)
	}).Start()
	com.NewThreadWithID("[Micro Chain Keep Alive]", c.KeepAlive, func(err interface{}) {
		panic(err)
	}).Start()

	com.NewThreadWithID("[Cmd Service Thread]", func(c chan struct{}) {
		StartCmdService()
	}, func(err interface{}) {
		panic(err)
	}).Start()

	done := make(chan bool, 1)
	go waitSignal(done)
	<-done
}

func waitSignal(done chan bool) {
	pid := strconv.Itoa(os.Getpid())
	fmt.Printf("\n>>>>>>>>>>miner start at pid(%s)<<<<<<<<<<\n", pid)
	if err := ioutil.WriteFile(node.PathSetting.PidPath, []byte(pid), 0644); err != nil {
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

type cmdService struct{}

func StartCmdService() {
	address := net.JoinHostPort("127.0.0.1", node.CMDServicePort)
	l, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	cmdServer := grpc.NewServer()

	pbs.RegisterCmdServiceServer(cmdServer, &cmdService{})

	reflection.Register(cmdServer)
	if err := cmdServer.Serve(l); err != nil {
		panic(err)
	}
}

func DialToCmdService() pbs.CmdServiceClient {
	var address = "127.0.0.1:" + node.CMDServicePort
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := pbs.NewCmdServiceClient(conn)

	return client
}
