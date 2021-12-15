package api

import (
	"encoding/json"
	"github.com/redeslab/go-miner/webserver/session"
	"io/ioutil"
	"net/http"
)

type Request struct {
	AccessToken string      `json:"access_token"`
	Data        interface{} `json:"data,omitempty"`
}

const (
	Success      int = 0
	ReadBodyErr  int = 1
	UnMarshalErr int = 2
	SessErr      int = 3
	NotAPost     int = 4
)

type Response struct {
	ResultCode int         `json:"result_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

func doRequest(r *http.Request, v interface{}) (*Request, *Response) {
	resp := &Response{}

	if r.Method != "POST" {
		resp.Message = "not a post request"
		resp.ResultCode = NotAPost
		return nil, resp
	}

	var (
		content []byte
		err     error
	)
	if content, err = ioutil.ReadAll(r.Body); err != nil {
		resp.Message = "read http body error"
		resp.ResultCode = ReadBodyErr
		return nil, resp
	}

	req := &Request{}
	req.Data = v
	err = json.Unmarshal(content, req)
	if err != nil {
		resp.ResultCode = UnMarshalErr
		resp.Message = "not a correct json"
		return nil, resp
	}

	if !session.IsValidBase58(req.AccessToken) {
		resp.ResultCode = SessErr
		resp.Message = "not a correct session"
		return nil, resp
	}

	resp.ResultCode = Success
	resp.Message = "success"

	return req, resp
}
