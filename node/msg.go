package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hyperorchidlab/go-miner-pool/account"
	"github.com/hyperorchidlab/go-miner-pool/microchain"
	"github.com/hyperorchidlab/go-miner-pool/network"
)

const (
	MsgDeliverMicroTx int = iota
	MsgSyncMicroTx
	MsgPingTest
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

type MsgReq struct {
	Typ int                 `json:"typ"`
	SMT *SyncMicroTx        `json:"smt,omitempty"`
	TX  *microchain.MicroTX `json:"tx,omitempty"`
	PT  *PingTest           `json:"pt,omitempty"`
}

type SyncMicroTx struct {
	User common.Address `json:"user"`
}

type MsgAck struct {
	Typ  int         `json:"typ"`
	Code int         `json:"code"` //0 success 1 failure
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
