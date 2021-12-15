package session

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/redeslab/go-miner/util"
	"github.com/redeslab/go-miner/webserver/constant"
	"math/rand"
	"sync"
	"time"
)

type SessionDesc struct {
	IsVerify       bool
	LastAccessTime int64
}

var (
	sessionTimeout int = 600 //10 minutes
	quit           chan struct{}
	wg             sync.WaitGroup
	accessSession  map[[constant.RandBytesCount]byte]*SessionDesc
)

func init() {
	quit = make(chan struct{})
	accessSession = make(map[[constant.RandBytesCount]byte]*SessionDesc)
}

func NewSession() ([constant.RandBytesCount]byte, *SessionDesc) {
	var randbytes [constant.RandBytesCount]byte

	for {
		rand.Seed(util.GetNowMsTime())
		n, err := rand.Read(randbytes[:])
		if err != nil || constant.RandBytesCount != n {
			continue
		}
		break
	}

	sd := &SessionDesc{LastAccessTime: util.GetNowMsTime()}

	accessSession[randbytes] = sd

	return randbytes, sd

}

func NewSession2() (string, *SessionDesc) {
	b, s := NewSession()

	return base58.Encode(b[:]), s
}

func IsSession(k [constant.RandBytesCount]byte) bool {
	if _, ok := accessSession[k]; !ok {
		return false
	}

	return true
}

func IsSessionBase58(k string) bool {
	kb := base58.Decode(k)

	if len(kb) != constant.RandBytesCount {
		return false
	}

	var key [constant.RandBytesCount]byte
	copy(key[:], kb)

	return IsSession(key)
}

func IsValid(k [constant.RandBytesCount]byte) bool {
	if s, ok := accessSession[k]; !ok {
		return false
	} else {
		if !s.IsVerify {
			return false
		}

		if util.GetNowMsTime()-s.LastAccessTime > (int64(sessionTimeout) * 1000) {
			return false
		}

		return true
	}
}

func IsValidBase58(k string) bool {
	kb := base58.Decode(k)

	if len(kb) != constant.RandBytesCount {
		return false
	}

	var key [constant.RandBytesCount]byte
	copy(key[:], kb)

	return IsValid(key)
}

func SessionActiveBase58(k string) {
	kb := base58.Decode(k)

	if len(kb) != constant.RandBytesCount {
		return
	}

	var key [constant.RandBytesCount]byte
	copy(key[:], kb)

	if v, ok := accessSession[key]; !ok {
		return
	} else {
		v.IsVerify = true
		v.LastAccessTime = util.GetNowMsTime()
	}

}

func StartTimeOut() {
	wg.Add(1)
	for {

		select {
		case <-quit:
			wg.Done()
			return
		default:

		}

		now := util.GetNowMsTime()

		var ks [][constant.RandBytesCount]byte
		for k, v := range accessSession {
			if now-v.LastAccessTime > (int64(sessionTimeout) * 1000) {
				ks = append(ks, k)
			}
		}

		for i := 0; i < len(ks); i++ {
			delete(accessSession, ks[i])
		}

		time.Sleep(time.Second)
	}
}

func StopTimeOut() {
	quit <- struct{}{}

	wg.Wait()
}
