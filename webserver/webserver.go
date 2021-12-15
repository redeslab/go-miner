package webserver

import (
	"context"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/redeslab/go-miner/node"
	"github.com/redeslab/go-miner/webserver/api"
	"github.com/redeslab/go-miner/webserver/session"
	"github.com/redeslab/go-miner/webserver/webfs"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"
)

var webserver *http.Server

func addHandle(mux *http.ServeMux, webpath string, handler http.Handler) {
	fmt.Println("add handle", webpath)
	mux.Handle(webpath, handler)
}

func StartWebDaemon() {
	mux := http.NewServeMux()

	sp := node.PathSetting

	addHandle(mux, path.Join("/api", sp.WebAuthPath, sp.WebAuthTokenPath), &api.WebAccessToken{})
	addHandle(mux, path.Join("/api", sp.WebAuthPath, sp.WebAuthVerifyPath), &api.SigVerification{})

	addHandle(mux, path.Join("/api", sp.WebMinerPath, sp.WebMinerDetails), &api.MinerInfo{})

	addHandle(mux, path.Join("/api", sp.WebUserPath, sp.WebUserCount), &api.UsersCountInMiner{})
	addHandle(mux, path.Join("/api", sp.WebUserPath, sp.WebUserInfo), &api.UsersInfoInMiner{})

	wfs := assetfs.AssetFS{Asset: webfs.Asset, AssetDir: webfs.AssetDir, AssetInfo: webfs.AssetInfo, Prefix: "webserver/html/dist"}

	mux.Handle("/", http.FileServer(&wfs))

	log.Println("Miner Management Web Server Start at:", node.WebPort)

	webserver = &http.Server{
		Addr:    ":" + strconv.Itoa(node.WebPort),
		Handler: mux,
	}

	go session.StartTimeOut()
	log.Println(webserver.ListenAndServe())
}

func StopWebDaemon() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	webserver.Shutdown(ctx)

	session.StopTimeOut()

	log.Println("Miner Management Web Server Stopped")
}
