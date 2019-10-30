package node

import (
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/common"
	com "github.com/hyperorchid/go-miner-pool/common"
	"github.com/hyperorchid/go-miner-pool/eth"
	"github.com/hyperorchid/go-miner-pool/network"
	"net"
	"sync"
)

var (
	mcInstance     *MicChain = nil
	mcOnce         sync.Once
	DBKeyMinerData = "_DB_KEY_MINER_DATA_FOR_POOL_%s"
)

type MicChain struct {
	conn      net.Conn
	database  *leveldb.DB
	minerData *MinerData
}

type MinerData struct {
	*eth.MinerData
	unSettled int64
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

	md, err := QueryMinerData(WInst().SubAddress())
	if err != nil {
		panic(err)
	}

	localMD := &MinerData{}
	mdKey := minerKey(md.PoolAddr)
	has, err := db.Has(mdKey, nil)

	if err != nil {
		panic(err)
	}
	if has {
		_ = com.GetJsonObj(db, minerKey(md.PoolAddr), localMD)
	} else {
		localMD = &MinerData{MinerData: md, unSettled: 0}
	}

	ip, err := network.BASInst().Query(md.PoolAddr[:])
	if err != nil {
		panic(err)
	}
	addr := net.JoinHostPort(string(ip), com.DefaultTxBroadPort)

	c, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	mc := &MicChain{
		conn:      c,
		database:  db,
		minerData: localMD,
	}

	return mc
}

func minerKey(poolAddr common.Address) []byte {
	return []byte(fmt.Sprintf(DBKeyMinerData, poolAddr))
}

func (mc *MicChain) sync(sig chan struct{}) {

}
