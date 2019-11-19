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
	minerData     *microchain.MinerTxData
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

	db, err := leveldb.OpenFile(SysConf.DBPath, &opts)
	if err != nil {
		panic(err)
	}

	minerID := WInst().SubAddress()
	md, err := QueryMinerData(minerID)
	if err != nil {
		panic(err)
	}
	chainLog.Notice("Sync miner data:", md.String())
	localMD := &microchain.MinerTxData{}
	mdKey := minerKey(minerID, md.PoolAddr)
	if err := com.GetJsonObj(db, mdKey, localMD); err != nil {
		localMD = &microchain.MinerTxData{PackMined: big.NewInt(0), EthData: md}
	}

	ntAddr, err := basc.QueryBySrvIP(md.PoolAddr.Bytes(), SysConf.BAS)
	if err != nil {
		panic(err)
	}
	addr := net.JoinHostPort(string(ntAddr.NetAddr), com.ReceiptSyncPort)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	_ = c.SetDeadline(time.Now().Add(time.Second * 2))
	conn := &network.JsonConn{Conn: c}

	syn := &microchain.ReceiptSync{
		Typ: microchain.ReceiptSyncTypeMiner,
		MR: &microchain.MinerReceipt{
			MinerTxData: localMD,
		},
	}
	syn.MR.Sig = WInst().SignJSONSub(syn.MR.MinerTxData)
	if err := conn.WriteJsonMsg(syn); err != nil {
		panic(err)
	}

	ack := &microchain.MinerTxData{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		panic(err)
	}

	fmt.Println(ack.String())

	if localMD.LastMicNonce < ack.LastMicNonce {
		log.Warn("account isn't same and corrected", localMD.String(), ack.String())
		localMD.PackMined = ack.PackMined
		localMD.LastMicNonce = ack.LastMicNonce
		localMD.EthData = ack.EthData
		_ = com.SaveJsonObj(db, mdKey, localMD)
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
