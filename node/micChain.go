package node

import (
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/op/go-logging"
	basc "github.com/redeslab/BAS/client"
	com "github.com/redeslab/go-miner-pool/common"
	"github.com/redeslab/go-miner-pool/microchain"
	"github.com/redeslab/go-miner-pool/network"
	"net"
	"sync"
	"time"
)

var (
	mcInstance  *MicChain = nil
	mcOnce      sync.Once
	chainLog, _ = logging.GetLogger("chain")
)

type MicChain struct {
	Sign          *microchain.ReceiptSync
	conn          *network.JsonConn
	database      *leveldb.DB
	BucketManager BucketManager
}

func Chain() *MicChain {
	mcOnce.Do(func() {
		mcInstance = newChain()
	})
	return mcInstance
}

func newChain() *MicChain {
	opts := opt.Options{
		Strict:      opt.DefaultStrict,
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(PathSetting.DBPath, &opts)
	if err != nil {
		panic(err)
	}
	minerID := WInst().SubAddress()
	md, err := QueryMinerData(minerID)
	if err != nil {
		panic(err)
	}
	chainLog.Notice("Sync miner data:", md.String())
	ntAddr, err := basc.QueryBySrvIP(md.PoolAddr.Bytes(), SysConf.BAS)
	if err != nil {
		panic(err)
	}
	addr := &net.UDPAddr{IP: net.ParseIP(string(ntAddr.NetAddr)), Port: com.ReceiptSyncPort}
	c, err := net.DialTimeout("udp", addr.String(), time.Second*4)
	if err != nil {
		panic(err)
	}
	_ = c.SetDeadline(time.Now().Add(time.Second * 2))
	conn := &network.JsonConn{Conn: c}

	syn := &microchain.ReceiptSync{
		ReceiptQueryData: &microchain.ReceiptQueryData{
			Typ:       microchain.ReceiptSyncTypeMiner,
			QueryAddr: WInst().SubAddress().String(),
			PoolAddr:  md.PoolAddr,
		},
	}
	syn.Sig = WInst().SignJSONSub(syn.ReceiptQueryData)
	if err := conn.WriteJsonMsg(syn); err != nil {
		panic(err)
	}
	_ = conn.SetDeadline(time.Time{})

	mc := &MicChain{
		conn:     conn,
		database: db,
		Sign:     syn,
	}
	return mc
}

func (mc *MicChain) Sync(_ chan struct{}) {
	r := &microchain.Receipt{}
	for {
		if err := mc.conn.ReadJsonMsg(r); err != nil {
			panic(err)
		}
		chainLog.Notice(r.String())
		if err := mc.BucketManager.RechargeBucket(r); err != nil {
			log.Warn("recharge err:", err)
			continue
		}
		mc.saveReceipt(r)
	}
}
func (mc *MicChain) KeepAlive(_ chan struct{}) {
	for {
		select {
		case <-time.After(30 * time.Second):
			if err := mc.conn.WriteJsonMsg(mc.Sign); err != nil {
				panic(err) //TODO:: try to join pool again
			}
		}
	}
}

func (mc *MicChain) saveReceipt(_ *microchain.Receipt) {
	//TODO::make a Merckle tree
}
