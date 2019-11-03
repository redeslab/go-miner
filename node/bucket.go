package node

import (
	"fmt"
	"github.com/hyperorchid/go-miner-pool/microchain"
	"sync"
)

const (
	InitBucketSize = 1 << 22
) //4M
var (
	ErrNoPacketBalance = fmt.Errorf("need to recharge for this mienr")
)

type BucketManager interface {
	RechargeBucket(*microchain.Receipt) error
}
type Bucket struct {
	sync.RWMutex
	token int
}

func NewBucket() *Bucket {
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
	if b.token <= 0 {
		return ErrNoPacketBalance
	}
	return nil
}

func (b *Bucket) Recharge(no int) {
	b.Lock()
	defer b.Unlock()
	b.token += no
}
