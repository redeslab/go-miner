package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/util"
	"github.com/ethereum/go-ethereum/common"
	com "github.com/hyperorchidlab/go-miner-pool/common"
	"github.com/hyperorchidlab/go-miner-pool/microchain"
	"strings"

	"log"
	"math/big"
	"sync"
)

type UserAccount struct {
	TokenBalance   *big.Int
	TrafficBalance *big.Int
	TotalTraffic   *big.Int

	UptoPoolTraffic *big.Int
	MinerCredit     *big.Int

	PoolRefused bool
}

func (ua *UserAccount) dup() *UserAccount {
	return &UserAccount{
		TokenBalance:    ua.TokenBalance,
		TrafficBalance:  ua.TrafficBalance,
		TotalTraffic:    ua.TotalTraffic,
		UptoPoolTraffic: ua.UptoPoolTraffic,
		MinerCredit:     ua.MinerCredit,
		PoolRefused:     ua.PoolRefused,
	}
}

type UserAccountMgmt struct {
	poolAddr common.Address
	users    map[common.Address]*UserAccount
	lock     map[common.Address]sync.RWMutex
	dblock   map[string]sync.RWMutex
	database *leveldb.DB
}

const (
	DBUserMicroTXHead string = "DBUserMicroTx_%s_%s"        //market pool
	DBUserMicroTxKey         = DBUserMicroTXHead + "_%s_%s" //user credit

	DBPoolMicroTxHead          string = "DBPoolMicroTx_%s_%s"        //market pool
	DBPoolMicroTxKey                  = DBPoolMicroTxHead + "_%s_%s" //user credit
	DBPoolMicroTxKeyPatternEnd        = "DBPoolMicroTx_0xffffffffffffffffffff"
)

func NewUserAccMgmt(db *leveldb.DB, pool common.Address) *UserAccountMgmt {
	return &UserAccountMgmt{
		poolAddr: pool,
		users:    make(map[common.Address]*UserAccount),
		lock:     make(map[common.Address]sync.RWMutex),
		dblock:   make(map[string]sync.RWMutex),
		database: db,
	}
}

func NewUserAccount() *UserAccount  {
	return &UserAccount{
		TokenBalance:&big.Int{},
		TrafficBalance:&big.Int{},
		TotalTraffic:&big.Int{},
		UptoPoolTraffic:&big.Int{},
		MinerCredit:&big.Int{},
	}
}

func (uam *UserAccountMgmt) dbMicroTxKeyGet(fmts string, user common.Address, credit *big.Int) string {
	return fmt.Sprintf(fmts, SysConf.MicroPaySys.String(), uam.poolAddr.String(), user.String(), credit.String())
}

func (uam *UserAccountMgmt) DBUserMicroTXKeyGet(user common.Address, credit *big.Int) string {
	return uam.dbMicroTxKeyGet(DBUserMicroTxKey, user, credit)
}

func (uam *UserAccountMgmt) DBPoolMicroTxKeyGet(user common.Address, credit *big.Int) string {
	return uam.dbMicroTxKeyGet(DBPoolMicroTxKey, user, credit)
}

func (uam *UserAccountMgmt) DBUserMicroTXKeyDerive(key string) (user common.Address, credit *big.Int, err error) {
	arr := strings.Split(key, "_")
	if len(arr) != 5 {
		return user, nil, errors.New("key error")
	}

	user = common.HexToAddress(arr[3])
	credit, _ = (&big.Int{}).SetString(arr[4], 10)

	return user, credit, nil

}

func (uam *UserAccountMgmt) DBPoolMicroTxKeyDerive(key string) (user common.Address, credit *big.Int, err error) {
	return uam.DBUserMicroTXKeyDerive(key)
}

func (uam *UserAccountMgmt) checkMicroTx(tx *microchain.MicroTX) bool {
	locker := uam.lock[tx.User]
	locker.RLock()
	defer locker.RUnlock()

	ua, ok := uam.users[tx.User]
	if !ok {
		fmt.Println("check microtx,1")
		return false
	}

	if ua.PoolRefused {
		fmt.Println("check microtx,2")
		return false
	}

	zamount := &big.Int{}
	zamount = zamount.Sub(tx.MinerCredit, ua.MinerCredit)
	if zamount.Cmp(tx.MinerAmount) != 0 {
		fmt.Println("check microtx,3")
		return false
	}

	if tx.UsedTraffic.Cmp(ua.TrafficBalance) > 0 {
		fmt.Println("check microtx,4")
		return false
	}
	return true
}

func (uam *UserAccountMgmt) updateByMicroTx(tx *microchain.MicroTX) {
	locker := uam.lock[tx.User]
	locker.Lock()
	defer locker.Unlock()

	ua, ok := uam.users[tx.User]
	if !ok {
		log.Print("unexpected error, not found user account")
		return
	}

	ua.TotalTraffic = tx.UsedTraffic
	ua.MinerCredit = tx.MinerCredit

}

func (uam *UserAccountMgmt) saveUserMinerMicroTx(tx *microchain.MinerMicroTx) error {
	key := uam.DBUserMicroTXKeyGet(tx.User, tx.MinerCredit)
	locker := uam.dblock[key]
	locker.Lock()
	defer locker.Unlock()

	return com.SaveJsonObj(uam.database, []byte(key), *tx)
}

func (uam *UserAccountMgmt) savePoolMinerMicroTx(tx *microchain.DBMicroTx) error {
	key := uam.DBPoolMicroTxKeyGet(tx.User, tx.MinerCredit)
	locker := uam.dblock[key]
	locker.Lock()
	defer locker.Unlock()

	return com.SaveJsonObj(uam.database, []byte(key), *tx)
}

func (uam *UserAccountMgmt) dbGetMinerMicroTx(tx *microchain.MicroTX) (*microchain.MinerMicroTx, error) {
	key := uam.DBUserMicroTXKeyGet(tx.User, tx.MinerCredit)
	locker := uam.dblock[key]
	locker.RLock()
	defer locker.RUnlock()

	dbtx := &microchain.MinerMicroTx{}

	err := com.GetJsonObj(uam.database, []byte(key), dbtx)

	return dbtx, err
}

func (uam *UserAccountMgmt) resetCredit(user common.Address, credit *big.Int) {
	locker := uam.lock[user]
	locker.Lock()
	defer locker.Unlock()

	ua, ok := uam.users[user]
	if !ok {
		ua = NewUserAccount()
		uam.users[user] = ua
	}
	ua.MinerCredit = credit
}

func (uam *UserAccountMgmt) resetFromPool(user common.Address, sua *microchain.SyncUA) {
	locker := uam.lock[user]
	locker.Lock()
	defer locker.Unlock()

	ua, ok := uam.users[user]
	if !ok {
		ua = NewUserAccount()
		uam.users[user] = ua
	}
	ua.TotalTraffic = sua.UsedTraffic
	ua.TokenBalance = sua.TokenBalance
	ua.TrafficBalance = sua.TrafficBalance
	ua.UptoPoolTraffic = sua.UsedTraffic
	ua.PoolRefused = false

}

func (uam *UserAccountMgmt) syncBalance(user common.Address, sua *microchain.SyncUA) {
	locker := uam.lock[user]
	locker.Lock()
	defer locker.Unlock()

	ua, ok := uam.users[user]
	if !ok {
		return
	}
	//ua.TotalTraffic = sua.UsedTraffic
	ua.TokenBalance = sua.TokenBalance
	ua.TrafficBalance = sua.TrafficBalance
}

func (uam *UserAccountMgmt) getUserAcc(user common.Address) *UserAccount {
	locker := uam.lock[user]
	locker.RLock()
	defer locker.RUnlock()

	ua, ok := uam.users[user]
	if !ok {
		return nil
	}

	return ua.dup()

}

func (uam *UserAccountMgmt) refuse(user common.Address) {
	locker := uam.lock[user]
	locker.Lock()
	defer locker.Unlock()

	ua, ok := uam.users[user]
	if !ok {
		return
	}

	ua.PoolRefused = true
}

func (uam *UserAccountMgmt) getLatestMicroTx(user common.Address) *microchain.DBMicroTx {

	ua := uam.getUserAcc(user)
	if ua == nil {
		return nil
	}

	key := uam.DBPoolMicroTxKeyGet(user, ua.UptoPoolTraffic)

	locker := uam.dblock[key]
	locker.RLock()
	defer locker.RUnlock()

	dbtx := &microchain.DBMicroTx{}

	err := com.GetJsonObj(uam.database, []byte(key), dbtx)
	if err != nil {
		return nil
	}

	return dbtx
}

func (uam *UserAccountMgmt) loadFromDB() {
	pattern := fmt.Sprintf(DBPoolMicroTxHead, SysConf.MicroPaySys.String(), uam.poolAddr.String())

	r := &util.Range{Start: []byte(pattern), Limit: []byte(DBPoolMicroTxKeyPatternEnd)}

	iter := uam.database.NewIterator(r, nil)
	for iter.Next() {
		user, _, _ := uam.DBPoolMicroTxKeyDerive(string(iter.Key()))
		var (
			ua *UserAccount
			ok bool
		)
		if ua, ok = uam.users[user]; !ok {
			ua = &UserAccount{}
			uam.users[user] = ua
		}

		dbtx := &microchain.DBMicroTx{}
		json.Unmarshal(iter.Value(), dbtx)

		ua.MinerCredit = dbtx.MinerCredit
		ua.TrafficBalance = dbtx.TrafficBalance
		ua.TokenBalance = dbtx.TokenBalance
		ua.TotalTraffic = dbtx.UsedTraffic
		ua.UptoPoolTraffic = dbtx.MinerCredit
	}

}
