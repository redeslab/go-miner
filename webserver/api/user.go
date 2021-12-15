package api

import (
	"encoding/json"
	"github.com/redeslab/go-miner/node"
	"github.com/redeslab/pirate_contract/util"
	util2 "github.com/redeslab/pirate_contract/util"
	"net/http"
)

type UsersCountInMiner struct {
}

type MinerUserInfo struct {
	MainAddr       string `json:"main_addr"`
	HopBalance     string `json:"hop_balance"`
	TrafficBalance string `json:"traffic_balance"`
	TotalTraffic   string `json:"total_traffic"`
	MinerCredit    string `json:"miner_credit"`
	//TrafficDrawn   string `json:"traffic_drawn"`
}

func (ucm *UsersCountInMiner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var forceRefresh bool
	_, resp := doRequest(r, &forceRefresh)
	if resp.ResultCode > 0 {
		j, _ := json.Marshal(*resp)
		w.WriteHeader(200)
		w.Write(j)
		return
	}

	count := node.SrvNode().GetUserCount()

	resp.Data = &count

	j, _ := json.Marshal(*resp)

	w.WriteHeader(200)
	w.Write(j)
}

type UsersInfoInMiner struct {
}

type UserReqParam struct {
	PageSize     int  `json:"page_size"`
	PageNum      int  `json:"page_num"`
	ForceRefresh bool `json:"force_refresh"`
}

type UserListInfo struct {
	Count    int              `json:"count"`
	PageSize int              `json:"page_size"`
	PageNum  int              `json:"page_num"`
	Users    []*MinerUserInfo `json:"users"`
}

func (uim *UsersInfoInMiner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urp := &UserReqParam{}

	req, resp := doRequest(r, urp)
	if resp.ResultCode > 0 {
		j, _ := json.Marshal(*resp)
		w.WriteHeader(200)
		w.Write(j)
		return
	}

	param := req.Data.(*UserReqParam)
	data := &UserListInfo{
		PageNum:  param.PageNum,
		PageSize: param.PageSize,
	}
	resp.Data = data

	users := node.SrvNode().GetUsers()

	b, e := param.PageNum*param.PageSize, (param.PageNum+1)*param.PageSize
	if len(users) > b {
		if len(users) < e {
			e = len(users)
		}
		for i := b; i < e; i++ {
			uas := node.SrvNode().GetUserAccount(users[i])
			if uas == nil {
				continue
			}

			ud := &MinerUserInfo{
				MainAddr:       uas.UserAddress.String(),
				HopBalance:     util.Float2String(util.BalanceHuman(uas.TokenBalance), 4),
				TrafficBalance: util.Float2String(util2.TrafficGBytes(uas.TrafficBalance), 2),
				TotalTraffic:   util.Float2String(util2.TrafficGBytes(uas.TotalTraffic), 2),
				MinerCredit:    util.Float2String(util2.TrafficGBytes(uas.MinerCredit), 2),
			}
			if ud != nil {
				data.Count++
				data.Users = append(data.Users, ud)
			}
		}
	}

	j, _ := json.Marshal(*resp)
	w.WriteHeader(200)
	w.Write(j)
}
