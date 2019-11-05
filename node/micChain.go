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
	"net"
	"sync"
)

var (
	mcInstance     *MicChain = nil
	mcOnce         sync.Once
	DBKeyMinerData = "_DB_KEY_MINER_DATA_FOR_POOL_%s_%s"
)

type MicChain struct {
	conn          *network.JsonConn
	database      *leveldb.DB
	minerData     *MinerData
	BucketManager BucketManager
}

type MinerData struct {
	subAddr      account.ID
	poolAddr     common.Address
	PackMined    int64
	LastMicNonce int64
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

	localMD := &MinerData{}
	mdKey := minerKey(minerID, md.PoolAddr)
	has, err := db.Has(mdKey, nil)
	if err != nil {
		panic(err)
	}
	if has {
		_ = com.GetJsonObj(db, mdKey, localMD)
	} else {
		localMD = &MinerData{subAddr: minerID, poolAddr: md.PoolAddr}
	}

	ntAddr, err := network.BASInst().Query(md.PoolAddr.Bytes())
	if err != nil {
		panic(err)
	}
	addr := net.JoinHostPort(string(ntAddr.NetAddr), com.ReceiptSyncPort)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	conn := &network.JsonConn{Conn: c}
	reg := &microchain.ReaderReg{
		ReaderRegData: &microchain.ReaderRegData{
			SubAddr:   minerID,
			PackMined: localMD.PackMined,
		},
	}
	reg.Sig = WInst().SignJSONSub(reg)
	if err := conn.WriteJsonMsg(reg); err != nil {
		panic(err)
	}

	res := &microchain.ReaderRes{}
	if err := conn.ReadJsonMsg(res); err != nil {
		panic(err)
	}
	if !res.Verify() {
		panic("pool is not honest")
	}

	if localMD.poolAddr != res.Pool {
		panic("pool is not my manger")
	}

	if localMD.LastMicNonce < res.LastMicNonce {
		log.Warn("account isn't same and corrected")
		localMD.PackMined = res.PackMined
	}

	mc := &MicChain{
		conn:      conn,
		database:  db,
		minerData: localMD,
	}

	return mc
}

func minerKey(mid account.ID, poolAddr common.Address) []byte {
	return []byte(fmt.Sprintf(DBKeyMinerData, mid, poolAddr))
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
	_ = com.SaveJsonObj(mc.database, r.RKey(), r)
	mc.minerData.LastMicNonce = r.Nonce
	mc.minerData.PackMined += r.Amount
	_ = com.SaveJsonObj(mc.database, minerKey(r.Miner, r.To), mc.minerData)
}
