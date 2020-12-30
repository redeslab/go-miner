package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	basc "github.com/hyperorchidlab/BAS/client"
	"github.com/hyperorchidlab/go-miner-pool/account"
	com "github.com/hyperorchidlab/go-miner-pool/common"
	"github.com/hyperorchidlab/go-miner-pool/microchain"
	"github.com/hyperorchidlab/go-miner-pool/network"
	"github.com/hyperorchidlab/pirate_contract/config"
	"github.com/op/go-logging"
	"math/big"
	"net"
	"sync"
	"time"
)

var (
	instance   *Node = nil
	once       sync.Once
	nodeLog, _ = logging.GetLogger("node")
)

type Node struct {
	subAddr     account.ID
	poolAddr    common.Address
	payerAddr   common.Address
	poolNetAddr string
	poolConn    *net.UDPConn
	poolChan    chan *microchain.MinerMicroTx
	srvConn     net.Listener
	ctrlChan    *net.UDPConn
	buckets     *BucketMap
	database    *leveldb.DB
	uam         *UserAccountMgmt
	quit        chan struct{}
}

type NodeIns struct {
	SubAddr    account.ID
	PoolAddr   common.Address
	PayerAddr  common.Address
	Database   *leveldb.DB
	UAM        *UserAccountMgmt
}

func SrvNode() *Node {
	once.Do(func() {
		instance = newNode()
	})
	return instance
}

func newNode() *Node {
	sa := WInst().SubAddress()

	cfg := &config.PlatEthConfig{
		EthConfig: config.EthConfig{Market: SysConf.MicroPaySys, NetworkID: SysConf.NetworkID, EthApiUrl: SysConf.EthApiUrl, Token: SysConf.Token},
	}

	pool, payeraddr, err := GetPoolAddr(sa.ToArray(), cfg)
	if err != nil {
		panic(err)
	}

	opts := opt.Options{
		Strict:      opt.DefaultStrict,
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(PathSetting.DBPath, &opts)
	if err != nil {
		panic(err)
	}

	c, err := net.Listen("tcp", fmt.Sprintf(":%d", sa.ToServerPort()))
	if err != nil {
		panic(err)
	}
	p, err := net.ListenUDP("udp", &net.UDPAddr{Port: int(sa.ToServerPort())})
	if err != nil {
		panic(err)
	}

	bc := basc.NewBasCli(SysConf.BAS)
	fmt.Printf("%s\n", "===")
	fmt.Printf("%p\n", *pool)
	fmt.Printf("%s\n", "===")
	naddr, err := bc.Query((*pool)[:])
	if err != nil {
		panic(err)
	}
	ip := net.ParseIP(string(naddr.NetAddr))
	if ip.Equal(net.IPv4zero) {
		panic("pool ip address error:" + string(naddr.NetAddr))
	}

	uam := NewUserAccMgmt(db, *pool)
	uam.loadFromDB()

	n := &Node{
		subAddr:     sa,
		poolAddr:    *pool,
		payerAddr:   *payeraddr,
		poolNetAddr: string(naddr.NetAddr),
		poolChan:    make(chan *microchain.MinerMicroTx, 1024),
		srvConn:     c,
		ctrlChan:    p,
		buckets:     newBucketMap(),
		database:    db,
		uam:         uam,
		quit:        make(chan struct{}, 16),
	}

	com.NewThreadWithID("[report thread]", n.ReportTx, func(err interface{}) {
		panic(err)
	}).Start()

	com.NewThreadWithID("[UDP Test Thread]", n.CtrlService, func(err interface{}) {
		panic(err)
	}).Start()

	com.NewThreadWithID("[Buckets checker thread]", n.buckets.BucketTimer, func(err interface{}) {
		panic(err)
	}).Start()
	return n
}

func (n *Node) reportTx(tx *microchain.MinerMicroTx) (*microchain.PoolMicroTx, error) {
	if n.poolConn == nil {
		raddr := &net.UDPAddr{IP: net.ParseIP(n.poolNetAddr), Port: com.TxReceivePort}
		udpc, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			return nil, err
		}
		n.poolConn = udpc
	}

	fmt.Println("report tx 1:", tx.String())
	j, _ := json.Marshal(*tx)
	nw, err := n.poolConn.Write(j)
	if err != nil || nw != len(j) {
		n.poolConn.Close()
		n.poolConn = nil
		fmt.Println("report tx2:", err)
		return nil, err
	}

	ack := &microchain.PoolTxAck{}
	ptx := &microchain.PoolMicroTx{}
	ack.Data = ptx

	buf := make([]byte, 10240)
	n.poolConn.SetDeadline(time.Now().Add(time.Second * 2))
	nr, e := n.poolConn.Read(buf)
	if e != nil {
		n.poolConn.Close()
		n.poolConn = nil
		fmt.Println("report tx3:", err)
		return nil, e
	}
	n.poolConn.SetDeadline(time.Time{})

	err = json.Unmarshal(buf[:nr], ack)
	if err != nil {
		fmt.Println("report tx4:", err)
		return nil, err
	}

	if ack.Code == 0 {
		fmt.Println("report tx,get pool tx:", ptx.String())
		return ptx, nil
	}
	fmt.Println("report tx5:", ack.String())
	return nil, errors.New(ack.Msg)

}

func (n *Node) ReportTx(sig chan struct{}) {
	for {
		select {
		case tx := <-n.poolChan:
			ua := n.uam.getUserAcc(tx.User)
			if ua == nil {
				panic("unexpected no user account in mem")
			}
			if ptx, err := n.reportTx(tx); err == nil {
				dbtx := &microchain.DBMicroTx{TokenBalance: ua.TokenBalance, TrafficBalance: ua.TrafficBalance, PoolMicroTx: *ptx}
				if err := n.uam.savePoolMinerMicroTx(dbtx); err != nil {
					nodeLog.Warning("save dbtx error" + dbtx.String())
				}
			} else {
				n.uam.refuse(tx.User)
			}

		case <-n.quit:
			return
		}
	}
}

func (n *Node) ctrlChanRecv(req *MsgReq) *MsgAck {
	ack := &MsgAck{}
	ack.Typ = req.Typ
	ack.Msg = "failure"
	ack.Code = 1
	fmt.Println("Control Channel Receive:", req.String())
	switch req.Typ {
	case MsgDeliverMicroTx:
		if req.TX == nil {
			fmt.Println("1")
			return ack
		}
		if m, err := n.uam.dbGetMinerMicroTx(req.TX); err == nil {
			ack.Data = m
			ack.Msg = "success"
			ack.Code = 0
			fmt.Println("2")
			break
		}
		if b := n.uam.checkMicroTx(req.TX); !b {
			fmt.Println("3")
			return ack
		}
		var (
			sig []byte
			err error
		)
		if sig, err = WInst().SignJson(*req.TX); err != nil {
			fmt.Println("4")
			return ack
		}
		mtx := &microchain.MinerMicroTx{
			MinerSig: sig,
			MicroTX:  req.TX,
		}
		err = n.uam.saveUserMinerMicroTx(mtx)
		if err != nil {
			fmt.Println("5")
			return ack
		}

		fmt.Println("MinerMicroTx Save To DB", mtx.String())

		n.poolChan <- mtx
		n.uam.updateByMicroTx(req.TX)
		n.RechargeBucket(req.TX)
		ack.Data = mtx
		ack.Code = 0
		ack.Msg = "success"
	case MsgSyncMicroTx:
		if req.SMT == nil {
			return ack
		}

		tx, f, err := n.SyncMicro(req.SMT.User)
		if err != nil {
			fmt.Println("sync micro err", err, req.SMT.User.String())
			return ack
		}

		if f {
			fmt.Println("update ua by pool tx", tx.String())
			n.uam.resetCredit(req.SMT.User, tx.MinerCredit)
			ack.Data = tx.MinerMicroTx
		}

		sua, f, e := n.SyncUa(req.SMT.User)
		if e != nil {
			fmt.Println("sync ua err", req.SMT.User.String())
			return ack
		}

		if f {
			fmt.Println("begin reset ua from pool", sua.String())
			n.uam.resetFromPool(req.SMT.User, sua)
		}

		if ack.Data == nil {
			dbtx := n.uam.getLatestMicroTx(req.SMT.User)
			if dbtx != nil {
				ack.Data = dbtx.MinerMicroTx
			} else {
				ack.Code = 2
			}
		}
		if ack.Data != nil {
			ack.Code = 0
			ack.Msg = "success"
		}

		fmt.Println("answer to user", req.SMT.User.String(), ack.String())

	case MsgPingTest:
		ack.Code = 0
		ack.Msg = "success"
	}

	return ack
}

func (n *Node) CtrlService(sig chan struct{}) {
	for {
		buf := make([]byte, 10240)
		req := &MsgReq{}
		nr, addr, err := n.ctrlChan.ReadFrom(buf)
		if err != nil {
			log.Warn("control channel error ", err)
			continue
		}
		err = json.Unmarshal(buf[:nr], req)
		if err != nil {
			log.Warn("control channel bad request ", err)
			continue
		}

		data := n.ctrlChanRecv(req)
		j, _ := json.Marshal(*data)
		n.ctrlChan.WriteTo(j, addr)
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
	if n.poolConn != nil {
		n.poolConn.Close()
	}

	n.database.Close()
	close(n.quit)
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

func (n *Node) RechargeBucket(r *microchain.MicroTX) error {
	b := n.buckets.getBucket(r.User)
	if b == nil {
		return fmt.Errorf("no such user[%s] right now", r.User)
	}

	b.Recharge(int(r.MinerAmount.Int64()))
	return nil
}

func (n *Node) ShowUserBucket(user string) *Bucket {
	return n.buckets.getBucket(common.HexToAddress(user))

}

func (n *Node) dialPoolConn() (*net.TCPConn, error) {
	raddr := &net.TCPAddr{IP: net.ParseIP(string(n.poolNetAddr)), Port: com.SyncPort}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (n *Node) SyncMicro(user common.Address) (tx *microchain.DBMicroTx, find bool, err error) {
	conn, err := n.dialPoolConn()
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	lvconn := &network.LVConn{Conn: conn}
	jconn := &network.JsonConn{lvconn}

	sr := &microchain.SyncReq{}
	sr.Typ = microchain.RecoverMinerMicroTx
	sr.Miner = n.subAddr.ToArray()
	sr.UserAddr = user

	fmt.Println("begin to sync microtx from pool", sr.String())

	err = jconn.WriteJsonMsg(*sr)
	if err != nil {
		return nil, find, err
	}

	ptx := &microchain.DBMicroTx{}
	r := &microchain.SyncResp{}
	r.Data = ptx

	err = jconn.ReadJsonMsg(r)
	if err != nil {
		return nil, find, err
	}

	if r.Code == 0 {
		find = true
	}

	fmt.Println("receive ack microtx from pool", r.String())

	return ptx, find, nil
}

func (n *Node) SyncUa(user common.Address) (ua *microchain.SyncUA, find bool, err error) {
	conn, err := n.dialPoolConn()
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	lvconn := &network.LVConn{Conn: conn}
	jconn := &network.JsonConn{lvconn}

	sr := &microchain.SyncReq{}
	sr.Typ = microchain.SyncUserACC
	//sr.Miner = n.subAddr.ToArray()
	sr.UserAddr = user

	fmt.Println("Sync Ua from Pool", sr.String())

	err = jconn.WriteJsonMsg(*sr)
	if err != nil {
		fmt.Println("write to pool failed", user.String(), err)
		return nil, find, err
	}

	ua = &microchain.SyncUA{}

	r := &microchain.SyncResp{}
	r.Data = ua

	err = jconn.ReadJsonMsg(r)
	if err != nil {
		fmt.Println("read json error", err)
		return nil, find, err
	}

	fmt.Println("SyncUa resp:", r.String())

	if r.Code == 0 {
		find = true
	}

	return ua, find, nil
}

func (n *Node) GetNodeIns() *NodeIns {
	return &NodeIns{
		SubAddr:   n.subAddr,
		PoolAddr:  n.poolAddr,
		PayerAddr: n.payerAddr,
		Database:  n.database,
		UAM:       n.uam,
	}
}

func (n *Node) GetUserCount() int {
	return n.uam.GetUserCount()
}

func (n *Node) GetUsers() []common.Address {
	return n.uam.GetUsers()
}

func (n *Node) GetUserAccount(addr common.Address) *UserAccount {
	return n.uam.GetUserAccount(addr)
}

func (n *Node) GetMinerCredit() *big.Int {
	return n.uam.GetMinerCredit()
}