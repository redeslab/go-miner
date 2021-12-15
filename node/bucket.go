package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redeslab/go-miner-pool/microchain"
	"sync"
	"time"
)

const (
	InitBucketSize     = 1 << 24 //16M
	RechargePieceSize  = 1 << 22 //4M
	MaxLostRechargeReq = 16
	ConnectionBufSize  = 1 << 20
)

var (
	ErrNoPacketBalance = fmt.Errorf("need to recharge for this mienr")
)

type BucketManager interface {
	RechargeBucket(tx *microchain.MicroTX) error
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

func (bm *BucketMap) BucketTimer(sig chan struct{}) {
	for {
		select {
		case <-time.After(time.Minute * 15):
			now := time.Now()
			for key, val := range bm.Queue {
				if now.Sub(val.upTime) > time.Minute*30 {
					nodeLog.Noticef("bucket map remove :%s ", key.String())
					delete(bm.Queue, key)
				}
			}
		}
	}
}

func (bm *BucketMap) getBucket(addr common.Address) *Bucket {
	bm.RLock()
	defer bm.RUnlock()
	return bm.Queue[addr]
}

func (bm *BucketMap) delBucket(addr common.Address) {
	bm.Lock()
	defer bm.Unlock()
	delete(bm.Queue, addr)
}

func (bm *BucketMap) addPipe(addr common.Address) *Bucket {
	bm.Lock()
	defer bm.Unlock()
	if b, ok := bm.Queue[addr]; ok {
		return b
	}
	b := newBucket(len(bm.Queue))
	bm.Queue[addr] = b
	return b
}

type Bucket struct {
	BID int
	sync.RWMutex
	Token  int
	upTime time.Time
}

func newBucket(bid int) *Bucket {
	return &Bucket{
		BID:    bid,
		Token:  InitBucketSize,
		upTime: time.Now(),
	}
}

//Tips:: we just count the out put data
func (b *Bucket) ReadCount(no int) error {
	return nil
}

func (b *Bucket) WriteCount(no int) error {
	b.Lock()
	defer b.Unlock()
	b.upTime = time.Now()
	b.Token -= no
	//nodeLog.Debugf("bucket[%d] used:[%d] last:[%d]", b.BID, no, b.Token)
	if b.Token <= 0 {
		return ErrNoPacketBalance
	}
	return nil
}

func (b *Bucket) Recharge(no int) {
	b.Lock()
	defer b.Unlock()
	b.Token += no
	nodeLog.Noticef("bucket[%d] recharged:[%d]  now:[%d"+
		"]", b.BID, no, b.Token)
}
