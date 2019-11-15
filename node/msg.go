package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner-pool/network"
)

type SetupData struct {
	IV       network.Salt
	MainAddr common.Address
	SubAddr  account.ID
}

type SetupReq struct {
	Sig []byte
	*SetupData
}

type ProbeReq struct {
	Target string
}

func (sr *SetupReq) Verify() bool {
	return account.VerifyJsonSig(sr.MainAddr, sr.Sig, sr.SetupData)
}

type PingTest struct {
	PayLoad string
}
