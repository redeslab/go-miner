package node

import (
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchidlab/go-miner-pool/microchain"

	"log"
	"math/big"
	"sync"
)

type UserAccount struct {
	TokenBalance *big.Int
	TrafficBalance *big.Int
	TotalTraffic *big.Int

	MinerCredit *big.Int
}


type UserAccountMgmt struct {
	users map[common.Address]*UserAccount
	lock map[common.Address]sync.RWMutex
	dblock map[string]sync.RWMutex
	database *leveldb.DB
}

func NewUserAccMgmt(db *leveldb.DB) *UserAccountMgmt {
	return &UserAccountMgmt{
		users:make(map[common.Address]*UserAccount),
		lock:make(map[common.Address]sync.RWMutex),
		dblock: make(map[string]sync.RWMutex),
		database: db,
	}
}

func (uam *UserAccountMgmt)checkMicroTx(tx *microchain.MicroTX) bool  {
	locker:=uam.lock[tx.User]
	locker.RLock()
	defer locker.RUnlock()

	ua,ok:=uam.users[tx.User]
	if !ok{
		return false
	}

	zamount:=&big.Int{}
	zamount = zamount.Sub(tx.MinerCredit,ua.MinerCredit)
	if zamount.Cmp(tx.MinerAmount)!=0{
		return false
	}

	if tx.UsedTraffic.Cmp(ua.TrafficBalance) >0{
		return false
	}
	return true
}

func (uam *UserAccountMgmt)updateByMicroTx(tx *microchain.MicroTX)  {
	locker:=uam.lock[tx.User]
	locker.Lock()
	defer locker.Unlock()

	ua,ok:=uam.users[tx.User]
	if !ok{
		log.Print("unexpected error, not found user account")
		return
	}

	ua.TotalTraffic = tx.UsedTraffic
	ua.MinerCredit = tx.MinerCredit

}







