package node

import (
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hyperorchid/go-miner-pool/account"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner-pool/microchain"
	"github.com/hyperorchid/go-miner-pool/network"
	basc "github.com/hyperorchidlab/BAS/client"
	"github.com/op/go-logging"
	"math/big"
	"net"
	"sync"
	"time"
)

var (
	mcInstance     *MicChain = nil
	mcOnce         sync.Once
	DBKeyMinerData = "%s_DB_KEY_MINER_DATA_FOR_POOL_%s_%s"
	chainLog, _    = logging.GetLogger("chain")
)

type MicChain struct {
	conn          *network.JsonConn
	database      *leveldb.DB
	minerData     *MinerData
	BucketManager BucketManager
}

type MinerData struct {
	SubAddr      account.ID     `json:"address"`
	PoolAddr     common.Address `json:"pool"`
	PackMined    *big.Int       `json:"mined"`
	LastMicNonce int64          `json:"micNonce"`
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

	db, err := leveldb.OpenFile(SysConf.DBPath, &opts)
	if err != nil {
		panic(err)
	}
	minerID := WInst().SubAddress()
	md, err := QueryMinerData(minerID)
	if err != nil {
		panic(err)
	}
	if minerID != md.SubAddr {
		chainLog.Notice(md.String())
		panic("It's not my data")
	}
	chainLog.Notice("Sync miner data:", md.String())

	localMD := &MinerData{}
	mdKey := minerKey(minerID, md.PoolAddr)
	if err := com.GetJsonObj(db, mdKey, localMD); err != nil {
		localMD = &MinerData{SubAddr: minerID, PoolAddr: md.PoolAddr, PackMined: big.NewInt(0)}
		if err := com.SaveJsonObj(db, mdKey, localMD); err != nil {
			panic(err)
		}
	}

	ntAddr, err := basc.QueryBySrvIP(md.PoolAddr.Bytes(), SysConf.BAS)
	if err != nil {
		fmt.Println(md.String())
		panic(err)
	}
	addr := net.JoinHostPort(string(ntAddr.NetAddr), com.ReceiptSyncPort)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	_ = c.SetDeadline(time.Now().Add(time.Second * 2))
	conn := &network.JsonConn{Conn: c}
	syn := &microchain.MinerSyn{
		MinerSynData: &microchain.MinerSynData{
			SubAddr:      minerID,
			PackMined:    localMD.PackMined,
			LastMicNonce: localMD.LastMicNonce,
		},
	}
	syn.Sig = WInst().SignJSONSub(syn.MinerSynData)
	if err := conn.WriteJsonMsg(syn); err != nil {
		panic(err)
	}

	ack := &microchain.MinerAck{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		panic(err)
	}
	if !ack.Verify() {
		panic("pool is not honest")
	}
	chainLog.Notice("Sync miner data:", ack.String())

	if localMD.PoolAddr != ack.Pool {
		panic("pool is not my manger")
	}

	if localMD.LastMicNonce < ack.LastMicNonce {
		log.Warn("account isn't same and corrected")
		localMD.PackMined = ack.PackMined
	}

	mc := &MicChain{
		conn:      conn,
		database:  db,
		minerData: localMD,
	}
	_ = mc.conn.SetDeadline(time.Time{})
	return mc
}

func minerKey(mid account.ID, poolAddr common.Address) []byte {
	return []byte(fmt.Sprintf(DBKeyMinerData, SysConf.MicroPaySys.String(), mid.String(), poolAddr.String()))
}

func (mc *MicChain) Sync(sig chan struct{}) {
	r := &microchain.Receipt{}
	for {
		if err := mc.conn.ReadJsonMsg(r); err != nil {
			panic(err)
		}

		if mc.minerData.LastMicNonce >= r.Nonce {
			log.Warn("outdated receipt data")
			continue
		}

		if err := mc.BucketManager.RechargeBucket(r); err != nil {
			log.Warn("recharge err:", err)
			continue
		}
		mc.saveReceipt(r)

		select {
		case <-sig:
			log.Info("mic chain sync exit by other")
			return
		default:
		}
	}
}

func (mc *MicChain) saveReceipt(r *microchain.Receipt) {
	_ = com.SaveJsonObj(mc.database, r.RKey(SysConf.MicroPaySys), r)
	mc.minerData.LastMicNonce = r.Nonce
	mc.minerData.PackMined = mc.minerData.PackMined.Add(mc.minerData.PackMined, r.Amount)
	_ = com.SaveJsonObj(mc.database, minerKey(r.Miner, r.To), mc.minerData)
}
