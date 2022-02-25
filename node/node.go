package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/op/go-logging"
	basc "github.com/redeslab/BAS/client"
	"github.com/redeslab/BAS/dbSrv"
	"github.com/redeslab/go-miner-pool/account"
	com "github.com/redeslab/go-miner-pool/common"
	"github.com/redeslab/go-miner-pool/microchain"
	"github.com/redeslab/go-miner-pool/network"
	"github.com/redeslab/go-miner/bas"
	"github.com/redeslab/pirate_contract/config"
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
	SubAddr   account.ID
	PoolAddr  common.Address
	PayerAddr common.Address
	Database  *leveldb.DB
	UAM       *UserAccountMgmt
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
		EthConfig: config.EthConfig{Market: MinerSetting.MicroPaySys, NetworkID: MinerSetting.NetworkID, EthApiUrl: MinerSetting.EthApiUrl, Token: MinerSetting.Token},
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

	bc := basc.NewBasCli(MinerSetting.BAS)
	fmt.Printf("%s\n", "===")
	fmt.Printf("%s\n", *pool)
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

	if err := n.CheckVersion(); err != nil {
		panic(err)
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

	j, _ := json.Marshal(*tx)
	nw, err := n.poolConn.Write(j)
	if err != nil || nw != len(j) {
		n.poolConn.Close()
		n.poolConn = nil
		nodeLog.Warning("[reportTx] pool connection Write:=>", err)
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
		nodeLog.Warning("[reportTx] pool connection Read:=>", err)
		return nil, e
	}
	n.poolConn.SetDeadline(time.Time{})

	err = json.Unmarshal(buf[:nr], ack)
	if err != nil {
		nodeLog.Warning("[reportTx] Unmarshal:", err)
		return nil, err
	}

	if ack.Code == 0 {
		nodeLog.Debug("[reportTx] tx,get pool tx:", ptx.String())
		return ptx, nil
	}
	nodeLog.Debug("[reportTx] result:", ack.String())
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
				n.uam.setUpToTraffic(tx.User, ptx.MinerCredit)
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
	nodeLog.Debug("Control Channel Receive:", req.String())

	switch req.Typ {
	case MsgDeliverMicroTx:
		if req.TX == nil {
			nodeLog.Debug("1")
			return ack
		}
		if m, err := n.uam.dbGetMinerMicroTx(req.TX); err == nil {
			ack.Data = m
			ack.Msg = "success"
			ack.Code = 0
			break
		}
		if err := n.uam.checkMicroTx(req.TX); err != nil {
			nodeLog.Debug("", err, ack)
			return ack
		}
		var (
			sig []byte
			err error
		)
		sig = WInst().SignJSONSub(*req.TX)

		mtx := &microchain.MinerMicroTx{
			MinerSig: sig,
			MicroTX:  req.TX,
		}
		err = n.uam.saveUserMinerMicroTx(mtx)
		if err != nil {
			nodeLog.Debug("save user miner micro tx :=>", err)
			return ack
		}

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
			nodeLog.Warning("sync micro err", err, req.SMT.User.String())
			return ack
		}

		if f {
			nodeLog.Debug("update ua by pool tx", tx.String())
			n.uam.resetCredit(req.SMT.User, tx.MinerCredit)
			ack.Data = tx.MinerMicroTx
		}

		sua, f, e := n.SyncUa(req.SMT.User)
		if e != nil {
			nodeLog.Warning("sync ua err", req.SMT.User.String())
			return ack
		}

		if f {
			nodeLog.Debug("begin reset ua from pool", sua.String())
			n.uam.resetFromPool(req.SMT.User, sua)
		}

		mtx := n.uam.getLastestMicroTx(req.SMT.User)
		if mtx != nil && tx != nil && mtx.MinerCredit.Cmp(tx.MinerCredit) > 0 {
			ack.Data = mtx
		}

		if ack.Data == nil {
			ack.Code = 2
			ack.Msg = "no data"
		} else {
			ack.Code = 0
			ack.Msg = "success"
		}

		nodeLog.Debug("answer to user", req.SMT.User.String(), ack.String())

	case MsgPingTest:
		ack.Code = 0
		ack.Msg = "success"
	}

	return ack
}

func (n *Node) CtrlService(sig chan struct{}) {
	for {
		buf := make([]byte, 10240)

		nr, addr, err := n.ctrlChan.ReadFrom(buf)
		if err != nil {
			nodeLog.Warning("control channel error ", err)
			continue
		}
		go n.ctrlMsg(buf[:nr], addr)
	}
}

func (n *Node) ctrlMsg(buf []byte, addr net.Addr) {
	req := &MsgReq{}
	err := json.Unmarshal(buf, req)
	if err != nil {
		nodeLog.Warning("control channel bad request ", err)
		return
	}
	nodeLog.Debug("CtrlService raw data:", string(buf))
	data := n.ctrlChanRecv(req)
	if j, ejson := json.Marshal(*data); ejson != nil {
		nodeLog.Debug("Marshal ctrlMsg data failed", data.String())
		return
	} else {
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
			_ = conn.SetDeadline(time.Now().Add(time.Second * 10))
		}).Start()
	}
}

func (n *Node) Stop() {
	_ = n.srvConn.Close()
	if n.poolConn != nil {
		_ = n.poolConn.Close()
	}

	_ = n.database.Close()
	close(n.quit)
}

func (n *Node) newWorker(conn net.Conn) {
	nodeLog.Debug("new conn:", conn.RemoteAddr().String())
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	lvConn := network.NewLVConn(conn)
	jsonConn := &network.JsonConn{Conn: lvConn}
	req := &SetupReq{}
	if err := jsonConn.ReadJsonMsg(req); err != nil {
		panic(fmt.Errorf("read setup msg err:%s", err))
	}

	if !req.Verify() {
		nodeLog.Warning(req.String())
		panic("request signature failed")
	}
	jsonConn.WriteAck(nil)

	var aesKey account.PipeCryptKey
	if err := account.GenerateAesKey(&aesKey, req.SubAddr.ToPubKey(), WInst().CryptKey()); err != nil {
		panic(fmt.Errorf("generate aes key err:%s", err))
	}
	aesConn, err := network.NewAesConn(lvConn, aesKey[:], req.IV)
	if err != nil {
		panic(fmt.Errorf("create aes connection err:%s", err))
	}
	jsonConn = &network.JsonConn{Conn: aesConn}
	prob := &ProbeReq{}
	if err := jsonConn.ReadJsonMsg(prob); err != nil {
		panic(fmt.Errorf("read probe msg err:%s", err))
	}

	tgtConn, err := net.Dial("tcp", prob.Target)
	if err != nil {
		panic(fmt.Errorf("dial target[%s] err:%s", prob.Target, err))
	}
	_ = tgtConn.(*net.TCPConn).SetKeepAlive(true)

	jsonConn.WriteAck(nil)

	b := n.buckets.addPipe(req.MainAddr)
	cConn := network.NewCounterConn(aesConn, b)

	var peerMaxPacketSize = prob.MaxPacketSize
	if peerMaxPacketSize == 0 {
		peerMaxPacketSize = ConnectionBufSize
	}
	nodeLog.Debugf("Setup pipe[bid=%d] for:[%s] from:%s with peer max size=%d",
		b.BID,
		prob.Target,
		cConn.RemoteAddr().String(),
		peerMaxPacketSize)

	com.NewThread(func(sig chan struct{}) {
		buffer := make([]byte, peerMaxPacketSize)
		for {
			no, err := cConn.Read(buffer)
			if no == 0 {
				nodeLog.Warning("read from client failed", err, no)
				return
			}
			_, err = tgtConn.Write(buffer[:no])
			if err != nil {
				nodeLog.Warning("write to target failed", err)
				return
			}
		}
	}, nil).Start()
	buffer := make([]byte, peerMaxPacketSize)
	for {
		no, err := tgtConn.Read(buffer)
		if no == 0 {
			nodeLog.Warning("Target->Proxy read err:", err)
			_ = tgtConn.SetDeadline(time.Now().Add(time.Second * 10))
			return
		}

		_, err = cConn.Write(buffer[:no])
		if err != nil {
			nodeLog.Warning("Proxy->Client write  err:", err, no)
			return
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
		return nil, false, err
	}

	defer conn.Close()

	lvconn := &network.LVConn{Conn: conn}
	jconn := &network.JsonConn{lvconn}

	sr := &microchain.SyncReq{}
	sr.Typ = microchain.RecoverMinerMicroTx
	sr.Miner = n.subAddr.ToArray()
	sr.UserAddr = user

	nodeLog.Debug("begin to sync micro tx from pool", sr.String())

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
	} else {
		ptx = nil
	}

	nodeLog.Debug("receive ack micro tx from pool", r.String())

	return ptx, find, nil
}

func (n *Node) SyncUa(user common.Address) (ua *microchain.SyncUA, find bool, err error) {
	conn, err := n.dialPoolConn()
	if err != nil {
		return nil, false, err
	}

	defer conn.Close()

	lvconn := &network.LVConn{Conn: conn}
	jconn := &network.JsonConn{lvconn}

	sr := &microchain.SyncReq{}
	sr.Typ = microchain.SyncUserACC
	sr.UserAddr = user

	nodeLog.Debug("Sync Ua from Pool", sr.String())

	err = jconn.WriteJsonMsg(*sr)
	if err != nil {
		nodeLog.Warning("write to pool failed", user.String(), err)
		return nil, find, err
	}

	ua = &microchain.SyncUA{}

	r := &microchain.SyncResp{}
	r.Data = ua

	err = jconn.ReadJsonMsg(r)
	if err != nil {
		nodeLog.Warning("read json error", err)
		return nil, find, err
	}

	nodeLog.Debug("SyncUa resp:", r.String())

	if r.Code == 0 {
		find = true
	} else {
		ua = nil
	}

	return ua, find, nil
}

func (n *Node) UserManagement() *UserAccountMgmt {
	return n.uam
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

func (n *Node) CheckVersion() error {
	cnt := 0
	for {
		if err := n.checkVersion(); err != nil {
			cnt++
			if cnt > 5 {
				return err
			}
		} else {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (n *Node) checkVersion() error {
	client := basc.NewBasCli(MinerSetting.BAS)
	ba := n.subAddr.String()

	fmt.Println("ba ----->", ba)

	ext, nw, err := client.QueryExtend([]byte(ba))
	if err != nil {
		if bascerr, ok := err.(*basc.BascErr); ok {
			if bascerr.Code == basc.NoItemErr {
				panic(err)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	extd := &bas.MinerExtendData{}
	err = json.Unmarshal([]byte(ext), extd)
	if err != nil {
		return err
	}

	if extd.Version == HopVersion && extd.PoolAddr == n.poolAddr.String() {
		return nil
	}

	extd.Version = HopVersion
	extd.PoolAddr = n.poolAddr.String()

	req := &dbSrv.RegRequest{
		BlockAddr: []byte(n.subAddr.String()),
		SignData: dbSrv.SignData{
			NetworkAddr: nw,
			ExtData:     extd.Marshal(),
		},
	}

	req.Sig = WInst().SignJSONSub(req.SignData)
	if err := basc.RegisterBySrvIP(req, MinerSetting.BAS); err != nil {
		return err
	}

	return nil
}
