package node

import (
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/hyperorchid/go-miner-pool/eth"
	"sync"
)



var (
	mcInstance *MicChain = nil
	mcOnce     sync.Once
)

type MicChain struct {
	database *leveldb.DB
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

func newChain() *MicChain{

	opts := opt.Options{
		Strict:      opt.DefaultStrict,
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(SysConf.DBPath, &opts)
	if err != nil {
		panic(err)
	}

	md := QueryMyPool()

	mc := &MicChain{
		database:db,
	}

	return mc
}