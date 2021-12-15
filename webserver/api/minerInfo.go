package api

import (
	"encoding/json"
	basc "github.com/redeslab/BAS/client"
	"github.com/redeslab/go-miner/bas"
	"github.com/redeslab/go-miner/node"
	"github.com/redeslab/go-miner/webserver/util"
	"github.com/redeslab/pirate_contract/cabinet"
	"github.com/redeslab/pirate_contract/storageService"
	util2 "github.com/redeslab/pirate_contract/util"
	"math/big"
	"net/http"
)

type MinerInfo struct {
}

type MinerDesc struct {
	PoolAddr   string `json:"pool_addr"`
	SubAddr    string `json:"sub_addr"`
	PayerAddr  string `json:"payer_addr"`
	HopBalance string `json:"hop_balance"`
	Traffic    string `json:"traffic"`
	GTN        string `json:"gtn"`
	Zone       string `json:"zone"`
	Ip         string `json:"ip"`

	//UserList   [][32]byte `json:"user_list"`
}

func GetMinerDetail(ni *node.NodeIns) *MinerDesc {
	//mss := microchain.ChainInst().GetAllMiners()
	miner := ni.SubAddr.String()

	basip := node.MinerSetting.BAS
	basclient := basc.NewBasCli(basip)
	ext, _, err := basclient.QueryExtend([]byte(miner))

	extdata := &bas.MinerExtendData{}
	err = json.Unmarshal([]byte(ext), extdata)

	var ethset *cabinet.PirateEthSetting
	ethset, err = storageService.GetCacheSetting()

	var gtn *big.Int
	if err == nil {
		gtn = ethset.MinerDeposit
	}

	t := node.SrvNode().GetMinerCredit()
	if t == nil {
		t = &big.Int{}
	}

	h, err := storageService.Traffic2Balance(t)
	if err != nil {
		h = &big.Int{}
	}

	return &MinerDesc{
		PoolAddr:   ni.PoolAddr.String(),
		PayerAddr:  ni.PayerAddr.String(),
		SubAddr:    miner,
		Zone:       extdata.Location,
		Ip:         basip,
		GTN:        util.Float2String(util.BalanceHuman(gtn), 4),
		Traffic:    util.Float2String(util2.TrafficGBytes(t), 2),
		HopBalance: util.Float2String(util.BalanceHuman(h), 4),
	}
}

func (mi *MinerInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var forceRefresh bool
	_, resp := doRequest(r, &forceRefresh)
	if resp.ResultCode > 0 {
		j, _ := json.Marshal(*resp)
		w.WriteHeader(200)
		w.Write(j)
		return
	}

	ni := node.SrvNode().GetNodeIns()
	md := GetMinerDetail(ni)

	resp.Data = md
	j, _ := json.Marshal(*resp)

	w.WriteHeader(200)
	w.Write(j)
}
