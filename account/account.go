package account

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"sync"
)

type Account struct {
	sync.RWMutex
	Address ID
	Key     *Key
}

type SafeAccount struct {
	Version string `json:"version"`
	Address ID     `json:"address"`
	Cipher  string `json:"cipher"`
}

func (sa *SafeAccount) Unlock(auth string) (ed25519.PrivateKey, error) {
	pk := sa.Address.ToPubKey()
	aesKey, err := AESKey(pk[:KP.S], auth)
	if err != nil {
		return nil, err
	}

	cpTxt := make([]byte, len(sa.Cipher))
	copy(cpTxt, sa.Cipher)

	return Decrypt(aesKey, cpTxt)
}

func (acc *Account) IsEmpty() bool {
	return len(acc.Address) == 0
}

func (acc *Account) FormatShow() string {
	ret := fmt.Sprintf("\n**********************************************************************\n"+
		"\tNodeID:\t%s"+
		"\n**********************************************************************\n",
		acc.Address)

	return ret
}

func (acc *Account) UnlockAcc(password string) bool {
	pk := acc.Address.ToPubKey()

	aesKey, err := AESKey(pk[:KP.S], password) //scrypt.TxKey([]byte(password), k.PubKey[:KP.S], KP.N, KP.R, KP.P, KP.L)
	if err != nil {
		fmt.Println("error to generate aes key:->", err)
		return false
	}

	cpTxt := make([]byte, len(acc.Key.LockedKey))
	copy(cpTxt, acc.Key.LockedKey)

	raw, err := Decrypt(aesKey, cpTxt)
	if err != nil {
		fmt.Println("Unlock raw private key:->", err)
		return false
	}
	tmpPub, tmpPri := populateKey(raw)
	if !bytes.Equal(pk, tmpPub[:]) {
		fmt.Println("Unlock public failed")
		return false
	}

	acc.Key.PubKey = tmpPub
	acc.Key.PriKey = tmpPri
	return true
}

func (acc *Account) CreateAesKey(key *PipeCryptKey, peerAddr string) error {

	id, err := ConvertToID(peerAddr)
	if err != nil {
		return err
	}

	peerPub := id.ToPubKey()

	return acc.Key.GenerateAesKey(key, peerPub)
}

func (acc *Account) Sign(data []byte) []byte {
	return ed25519.Sign(acc.Key.PriKey, data)
}

func (acc *Account) Verify(data, sig []byte) bool {
	return ed25519.Verify(acc.Key.PubKey, data, sig)
}

func CheckID(address string) bool {
	if len(address) <= len(AccPrefix) {
		return false
	}

	id := ID(address)
	pk := id.ToPubKey()
	if len(pk) != ed25519.PublicKeySize {
		return false
	}

	return true
}

func AccFromString(addr, cipher, password string) (*Account, error) {

	address, err := ConvertToID(addr)
	if err != nil {
		return nil, err
	}

	acc := &Account{
		Address: address,
		Key: &Key{
			LockedKey: base58.Decode(cipher),
		},
	}

	if ok := acc.UnlockAcc(password); !ok {
		return nil, fmt.Errorf("unlock account failed")
	}

	return acc, nil
}
