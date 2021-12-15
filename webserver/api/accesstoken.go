package api

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/redeslab/go-miner/webserver/session"
	"net/http"
)

type WebAccessToken struct {
}

func (wat *WebAccessToken) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	randbytes, _ := session.NewSession()

	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", base58.Encode(randbytes[:]))
}
