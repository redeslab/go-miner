package node

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/redeslab/go-miner-pool/account"
	com "github.com/redeslab/go-miner-pool/common"
	"github.com/redeslab/go-miner-pool/microchain"
	"github.com/redeslab/go-miner-pool/network"
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

	com.NewThreadWithID("[Buckets checker thread]", n.buckets.BucketTimer, func(err interface{}) {
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
			_ = conn.Close()
		}).Start()
	}
}

func (n *Node) Stop() {
	_ = n.srvConn.Close()
}

const BUFFER_SIZE = 1 << 20

func (n *Node) newWorker(conn net.Conn) {
	log.Debug("new conn:", conn.RemoteAddr().String())
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	lvConn := network.NewLVConn(conn)
	jsonConn := &network.JsonConn{Conn: lvConn}
	req := &SetupReq{}
	if err := jsonConn.ReadJsonMsg(req); err != nil {
		panic(err)
	}

	if !req.Verify() {
		nodeLog.Warning(req.String())
		panic("request signature failed")
	}
	jsonConn.WriteAck(nil)

	var aesKey account.PipeCryptKey
	if err := account.GenerateAesKey(&aesKey, req.SubAddr.ToPubKey(), WInst().CryptKey()); err != nil {
		panic(err)
	}
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

	b := n.buckets.addPipe(req.MainAddr)
	cConn := network.NewCounterConn(aesConn, b)

	nodeLog.Debugf("Setup pipe[bid=%d] for:[%s] from:%s", b.BID, prob.Target, cConn.RemoteAddr().String())
	com.NewThread(func(sig chan struct{}) {
		buffer := make([]byte, ConnectionBufSize)
		for {
			no, err := cConn.Read(buffer)
			if err != nil && no == 0 {
				//nodeLog.Noticef("Client->Proxy read err:%s", err)
				panic(err)
			}
			_, err = tgtConn.Write(buffer[:no])
			if err != nil {
				//nodeLog.Noticef("Proxy->Target write err:%s", err)
				panic(err)
			}
		}
	}, func(err interface{}) {
		_ = tgtConn.Close()
	}).Start()
	buffer := make([]byte, ConnectionBufSize)
	for {
		no, err := tgtConn.Read(buffer)
		if err != nil && no == 0 {
			//nodeLog.Noticef("Target->Proxy read err:%s", err)
			panic(err)
		}
		_, err = cConn.Write(buffer[:no])
		if err != nil {
			//nodeLog.Noticef("Proxy->Client read err:%s", err)
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

func (n *Node) ShowUserBucket(user string) *Bucket {
	return n.buckets.getBucket(common.HexToAddress(user))

}
