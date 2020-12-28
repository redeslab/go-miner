package node

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperorchidlab/pirate_contract/config"
)

//
//func connect() (*generated.MicroPaySystem, error) {
//	conn, err := ethclient.Dial(SysConf.EthApiUrl)
//	if err != nil {
//		return nil, err
//	}
//	return generated.NewMicroPaySystem(SysConf.MicroPaySys, conn)
//}
//
//func tokenConn() (*ethclient.Client, *generated.Token, error) {
//	conn, err := ethclient.Dial(SysConf.EthApiUrl)
//	if err != nil {
//		return nil, nil, err
//	}
//	token, err := generated.NewToken(SysConf.Token, conn)
//	return conn, token, err
//}
//
//func QueryMinerData(subAddr account.ID) (*eth.MinerData, error) {
//	conn, err := connect()
//	if err != nil {
//		return nil, err
//	}
//
//	md, err := conn.MinerData(nil, subAddr.ToArray())
//	if err != nil {
//		return nil, err
//	}
//
//	miner := &eth.MinerData{
//		ID:        md.ID.Int64(),
//		PoolAddr:  md.PoolAddr,
//		PayerAddr: md.Payer,
//		SubAddr:   account.ConvertToID2(md.SubAddr[:]),
//		GTN:       md.GTN,
//		Zone:      string(md.Zone[:]),
//	}
//
//	return miner, nil
//}
//
//

func GetPoolAddr(miner [32]byte, cfg *config.PlatEthConfig) (addr *common.Address, err error) {
	if cfg == nil {
		return nil, errors.New("eth config error")
	}

	mc, err := cfg.NewClient()
	if err != nil {
		return nil, err
	}
	defer mc.Close()

	var ms [][32]byte
	ms = append(ms, miner)

	iter, err := mc.FilterMinerEvent(nil, ms, nil)
	if err != nil {
		return nil, err
	}

	var pool *common.Address

	for iter.Next() {
		ev := iter.Event
		if ev.EventType == 0 {
			pool = &ev.Addr1
		}

		if ev.EventType == 1 {
			pool = &ev.Addr2
		}

		if ev.EventType == 2 {
			pool = nil
		}
	}

	return pool, nil

}

//func QueryMinerData2(subAddr account.ID) (*eth.MinerData, error)  {
//
//}
