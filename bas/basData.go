package bas

import "encoding/json"

type MinerExtendData struct {
	MainAddr string `json:"main_addr"`
	HopAddr  string `json:"hop_addr"`
	Location string `json:"location"`
	PoolAddr string `json:"pool_addr,omitempty"`
	Version  string `json:"version,omitempty"`
}

func (med *MinerExtendData) Marshal() string {
	j, _ := json.Marshal(*med)
	return string(j)
}

func ToMinerExtData(s string) (*MinerExtendData, error) {
	med := &MinerExtendData{}

	err := json.Unmarshal([]byte(s), med)

	if err != nil {
		return nil, err
	}

	return med, nil
}
