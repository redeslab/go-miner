package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner-pool/network"
)

type SetupData struct {
	Target   string
	MainAddr common.Address
	SubAddr  account.ID
}

type SetupReq struct {
	IV  network.Salt
	Sig []byte
	*SetupData
}

func (sr *SetupReq) Verify() bool {
	return account.VerifyJsonSig(sr.MainAddr, sr.Sig, sr.SetupData)
}
