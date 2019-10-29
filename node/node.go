package node

import (
	"fmt"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner-pool/network"
	"io"
	"net"
	"sync"
)

var (
	instance *Node = nil
	once     sync.Once
)

type Node struct {
	wallet  account.Wallet
	subAddr account.ID
	srvConn net.Listener
}

func Inst() *Node {
	once.Do(func() {
		instance = newNode()
	})
	return instance
}

func newNode() *Node {

	w, err := account.LoadWallet(SysConf.WalletPath)
	if err != nil {
		panic(err)
	}

	sa := w.SubAddress()
	c, err := net.Listen("tcp", fmt.Sprintf(":%d", sa.ToServerPort()))
	if err != nil {
		panic(err)
	}

	n := &Node{
		wallet:  w,
		subAddr: sa,
		srvConn: c,
	}
	return n
}

func (n *Node) Mining() {
	for {
		conn, err := n.srvConn.Accept()
		if err != nil {
			panic(err)
		}
		go n.newWorker(conn)
	}
}

func (n *Node) Stop() {

}

func (n *Node) newWorker(conn net.Conn) {
	defer conn.Close()
	jsonConn := &network.JsonConn{Conn: conn}
	req := &SetupReq{}
	if err := jsonConn.ReadJsonMsg(req); err != nil {
		return
	}

	if !req.Verify() {
		return
	}
	jsonConn.WriteAck(nil)

	var aesKey account.PipeCryptKey
	if err := account.GenerateAesKey(&aesKey, req.SubAddr.ToPubKey(), n.wallet.CryptKey()); err != nil {
		return
	}
	lvConn := network.NewLVConn(conn)
	aesConn, err := network.NewAesConn(lvConn, aesKey[:], req.IV)
	if err != nil {
		return
	}
	jsonConn = &network.JsonConn{Conn: aesConn}
	prob := &ProbeReq{}
	if err := jsonConn.ReadJsonMsg(prob); err != nil {
		return
	}

	outSrvConn, err := net.Dial("tcp", prob.Target)
	if err != nil {
		return
	}
	inSrvConn := network.NewCounterConn(aesConn, n)
	io.Copy(inSrvConn, outSrvConn)
	io.Copy(outSrvConn, inSrvConn)
}

func (n *Node) ReadCount(no int) {

}

func (n *Node) WriteCount(no int) {

}
