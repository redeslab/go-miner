package node

import (
	"github.com/hyperorchid/go-miner-pool/account"
	"sync"
)

var (
	wInstance *MinerWallet = nil
	wOnce     sync.Once
)

type MinerWallet struct {
	account.Wallet
}

func WInst() *MinerWallet {
	once.Do(func() {
		wInstance = loadWallet()
	})
	return wInstance
}
func loadWallet() *MinerWallet {
	w, err := account.LoadWallet(SysConf.WalletPath)
	if err != nil {
		panic(err)
	}
	return &MinerWallet{Wallet: w}
}
