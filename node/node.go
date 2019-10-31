package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchid/go-miner-pool/account"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner-pool/network"
	"io"
	"net"
	"sync"
)

const (
	InitBucketSize = 1 << 22
) //4M

var (
	instance *Node = nil
	once     sync.Once
)

type Node struct {
	subAddr account.ID
	srvConn net.Listener
}

type userAccount struct {
	userAddr common.Address
}

type PipeJoiner struct {
	client net.Conn
	server net.Conn
}

func Inst() *Node {
	once.Do(func() {
		instance = newNode()
	})
	return instance
}

func newNode() *Node {
	sa := WInst().SubAddress()
	c, err := net.Listen("tcp", fmt.Sprintf(":%d", sa.ToServerPort()))
	if err != nil {
		panic(err)
	}

	n := &Node{
		subAddr: sa,
		srvConn: c,
	}
	return n
}

func (n *Node) Init() {
	//query eth for my pool
	//connect to pool and keep alive
	//sync all users under this pool
	//syncing version of user data
	//keep same of account between miner and pool
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
	if err := account.GenerateAesKey(&aesKey, req.SubAddr.ToPubKey(), WInst().CryptKey()); err != nil {
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

	tgtConn, err := net.Dial("tcp", prob.Target)
	if err != nil {
		return
	}
	cConn := network.NewCounterConn(aesConn, n)

	pj := &PipeJoiner{
		client: cConn,
		server: tgtConn,
	}

	com.NewThread(pj.PullFromServer, func(err interface{}) {
		_ = cConn.Close()
	}).Start()

	if _, err := io.Copy(pj.server, pj.client); err != nil {
		tgtConn.Close()
		return
	}
}

func (pj *PipeJoiner) PullFromServer(stopSig chan struct{}) {
	if _, err := io.Copy(pj.client, pj.server); err != nil {
		panic(err)
	}
}

func (n *Node) ReadCount(no int) {
	//TODO:: we just count the out put data
}

func (n *Node) WriteCount(no int) {
}
