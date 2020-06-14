package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func (sr *SetupReq) String() string {

	return fmt.Sprintf("\n@@@@@@@@@@@@@@@@@@@[Setup Request]@@@@@@@@@@@@@@@@@"+
		"\nSig:\t%s"+
		"\nIV:\t%s"+
		"\nMainAddr:\t%s"+
		"\nSubAddr:\t%s"+
		"\n@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@",
		hexutil.Encode(sr.Sig),
		hexutil.Encode(sr.IV[:]),
		sr.MainAddr.String(),
		sr.SubAddr.String())
}

type PingTest struct {
	PayLoad string
}
