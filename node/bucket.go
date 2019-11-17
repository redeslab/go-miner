package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchid/go-miner-pool/microchain"
	"sync"
)

const (
	InitBucketSize     = 1 << 24 //16M
	RechargePieceSize  = 1 << 22 //4M
	MaxLostRechargeReq = 4
)

var (
	ErrNoPacketBalance = fmt.Errorf("need to recharge for this mienr")
)

type BucketManager interface {
	RechargeBucket(*microchain.Receipt) error
}

type BucketMap struct {
	sync.RWMutex
	Queue map[common.Address]*Bucket
}

func newBucketMap() *BucketMap {
	return &BucketMap{
		Queue: make(map[common.Address]*Bucket),
	}
}

func (bm *BucketMap) newBucketItem(addr common.Address) *Bucket {
	b := newBucket()
	bm.Lock()
	defer bm.Unlock()
	bm.Queue[addr] = b
	return b
}
func (bm *BucketMap) getBucket(addr common.Address) *Bucket {
	bm.RLock()
	defer bm.RUnlock()
	return bm.Queue[addr]
}

type Bucket struct {
	sync.RWMutex
	token int
}

func newBucket() *Bucket {
	return &Bucket{
		token: InitBucketSize,
	}
}

//Tips:: we just count the out put data
func (b *Bucket) ReadCount(no int) error {
	return nil
}

func (b *Bucket) WriteCount(no int) error {
	b.Lock()
	defer b.Unlock()
	b.token -= no
	nodeLog.Debug("bucket used", no, " last:", b.token)
	if b.token <= 0 {
		return ErrNoPacketBalance
	}
	return nil
}

func (b *Bucket) Recharge(no int) {
	b.Lock()
	defer b.Unlock()
	b.token += no
	nodeLog.Notice("bucket recharged", no, " now:", b.token)
}
