package node

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/redeslab/go-miner-pool/account"
	"github.com/redeslab/go-miner-pool/microchain"
	"github.com/redeslab/go-miner-pool/network"
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
	Target        string `json:"Target"`
	MaxPacketSize int    `json:"MaxPacketSize,omitempty"`
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

func (pt *PingTest) String() string {
	if pt == nil {
		return "PingTest is nil"
	}
	return fmt.Sprintf("PayLoad: %s", pt.PayLoad)
}

type MsgReq struct {
	Typ int                 `json:"typ"`
	SMT *SyncMicroTx        `json:"smt,omitempty"`
	TX  *microchain.MicroTX `json:"tx,omitempty"`
	PT  *PingTest           `json:"pt,omitempty"`
}

func (mr *MsgReq) String() string {
	return fmt.Sprintf("type :%d\r\nSyncMicroTx: %s\r\nMicroTx: %sPingTest:%s\r\n",
		mr.Typ,
		mr.SMT.String(),
		mr.TX.String(),
		mr.PT.String())
}

type SyncMicroTx struct {
	User common.Address `json:"user"`
}

func (sm *SyncMicroTx) String() string {
	if sm == nil {
		return "SyncMicroTx is nil"
	}
	return fmt.Sprintf("User Address:%s", sm.User.String())
}

type MsgAck struct {
	Typ  int         `json:"typ"`
	Code int         `json:"code"` //0 success 1 failure
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (ack *MsgAck) String() string {
	if ack == nil {
		return "ack is nil"
	}
	j, err := json.Marshal(*ack)
	if err != nil {
		return err.Error()
	}

	return string(j)
}
