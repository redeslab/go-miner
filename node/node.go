package node

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hyperorchid/go-miner-pool/account"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner-pool/microchain"
	"github.com/hyperorchid/go-miner-pool/network"
	"github.com/op/go-logging"
	"net"
	"sync"
)

var (
	instance   *Node = nil
	once       sync.Once
	nodeLog, _ = logging.GetLogger("node")
)

type Node struct {
	subAddr account.ID
	srvConn net.Listener
	pingSrv *net.UDPConn
	buckets *BucketMap
}

func SrvNode() *Node {
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
	p, err := net.ListenUDP("udp", &net.UDPAddr{Port: int(sa.ToServerPort())})
	if err != nil {
		panic(err)
	}

	n := &Node{
		subAddr: sa,
		srvConn: c,
		pingSrv: p,
		buckets: newBucketMap(),
	}

	com.NewThreadWithID("[UDP Test Thread]", n.TestService, func(err interface{}) {
		panic(err)
	}).Start()
	return n
}

func (n *Node) TestService(sig chan struct{}) {
	buffer := make([]byte, 1024)
	for {
		_, a, e := n.pingSrv.ReadFromUDP(buffer)
		if e != nil {
			log.Warn("Test Ping:", e)
			continue
		}
		data, _ := json.Marshal(network.ACK{
			Success: true,
			Message: "",
		})
		_, _ = n.pingSrv.WriteToUDP(data, a)
	}
}

func (n *Node) Mining(sig chan struct{}) {
	defer n.srvConn.Close()
	for {
		conn, err := n.srvConn.Accept()
		if err != nil {
			panic(err)
		}

		com.NewThread(func(sig chan struct{}) {
			n.newWorker(conn)
		}, func(err interface{}) {
			nodeLog.Warning("Thread for proxy service exit:", conn.RemoteAddr().String(), err)
			conn.Close()
		}).Start()

		select {
		case <-sig:
			log.Info("mining exit by other")
			return
		default:
		}
	}
}

func (n *Node) Stop() {
	_ = n.srvConn.Close()
}

func (n *Node) newWorker(conn net.Conn) {
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	jsonConn := &network.JsonConn{Conn: conn}
	req := &SetupReq{}
	if err := jsonConn.ReadJsonMsg(req); err != nil {
		panic(err)
	}

	if !req.Verify() {
		panic("request signature failed")
	}
	jsonConn.WriteAck(nil)

	var aesKey account.PipeCryptKey
	if err := account.GenerateAesKey(&aesKey, req.SubAddr.ToPubKey(), WInst().CryptKey()); err != nil {
		panic(err)
	}
	lvConn := network.NewLVConn(conn)
	aesConn, err := network.NewAesConn(lvConn, aesKey[:], req.IV)
	if err != nil {
		panic(err)
	}
	jsonConn = &network.JsonConn{Conn: aesConn}
	prob := &ProbeReq{}
	if err := jsonConn.ReadJsonMsg(prob); err != nil {
		panic(err)
	}

	nodeLog.Debug("Request target:", prob.Target)
	tgtConn, err := net.Dial("tcp", prob.Target)
	if err != nil {
		panic(err)
	}
	_ = tgtConn.(*net.TCPConn).SetKeepAlive(true)

	jsonConn.WriteAck(nil)

	b := n.buckets.newBucketItem(req.MainAddr)
	cConn := network.NewCounterConn(aesConn, b)

	nodeLog.Noticef("Setup pipe for:[%s] from:%s", prob.Target, cConn.RemoteAddr().String())
	com.NewThread(func(sig chan struct{}) {
		buffer := make([]byte, 40960)
		for {
			no, err := cConn.Read(buffer)
			if err != nil && no == 0 {
				panic(err)
			}
			//fmt.Println("read from proxy lib->:", buffer[:no])
			_, err = tgtConn.Write(buffer[:no])
			if err != nil {
				panic(err)
			}
		}
	}, func(err interface{}) {
		nodeLog.Warning("service pull thread exit for:", err)
		_ = tgtConn.Close()
	}).Start()

	buffer := make([]byte, 40960)
	for {
		no, err := tgtConn.Read(buffer)
		if err != nil && no == 0 {
			panic(err)
		}
		//fmt.Println("read from target server->:", buffer[:no])
		_, err = cConn.Write(buffer[:no])
		if err != nil {
			panic(err)
		}
	}
}

func (n *Node) RechargeBucket(r *microchain.Receipt) error {
	b := n.buckets.getBucket(r.From)
	if b == nil {
		return fmt.Errorf("no such user[%s] right now", r.From)
	}

	b.Recharge(int(r.Amount.Int64()))
	return nil
}
