package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hyperorchid/go-miner-pool/account"
	"github.com/hyperorchid/go-miner-pool/eth"
	"github.com/hyperorchid/go-miner-pool/eth/generated"
)

func connect() (*generated.MicroPaySystem, error) {
	conn, err := ethclient.Dial(SysConf.EthApiUrl)
	if err != nil {
		return nil, err
	}
	return generated.NewMicroPaySystem(common.HexToAddress(SysConf.MicroPaySys), conn)
}

func tokenConn() (*ethclient.Client, *generated.Token, error) {
	conn, err := ethclient.Dial(SysConf.EthApiUrl)
	if err != nil {
		return nil, nil, err
	}
	token, err := generated.NewToken(common.HexToAddress(SysConf.Token), conn)
	return conn, token, err
}

func QueryMinerData(subAddr account.ID) (*eth.MinerData, error) {
	conn, err := connect()
	if err != nil {
		return nil, err
	}

	md, err := conn.MinerData(nil, subAddr.ToArray())
	if err != nil {
		return nil, err
	}

	miner := &eth.MinerData{
		ID:        md.ID.Int64(),
		PoolAddr:  md.PoolAddr,
		PayerAddr: md.Payer,
		SubAddr:   account.ConvertToID2(md.SubAddr[:]),
		GTN:       md.GuaranteedTokenNo.Int64(),
		Zone:      string(md.Zone[:]),
	}

	return miner, nil
}
