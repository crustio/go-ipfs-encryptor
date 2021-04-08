package crust

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ipfs/go-cid"
)

type sealResponse struct {
	Path       string `json:"path"`
	Message    string `json:"message"`
	StatusCode int64  `json:"status_code"`
}

type SWorker struct {
	url    string
	client http.Client
}

func NewSWorker(url string) *SWorker {
	client := http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	return &SWorker{url: url, client: client}
}

func (sw *SWorker) seal(ci cid.Cid, sessionKey string, isLink bool, value []byte) (bool, string, error) {
	// Generate request
	url := fmt.Sprintf("%s/storage/seal?cid=%s&session_key=%s&is_link=%t", sw.url, ci.String(), sessionKey, isLink)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(value))
	if err != nil {
		return false, "", err
	}

	// Request
	resp, err := sw.client.Do(req)
	if err != nil {
		return false, "", err
	}

	// Deal response
	if resp.StatusCode == 200 {
		returnBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return false, "", err
		}
		sealResp := &sealResponse{}
		err = json.Unmarshal(returnBody, sealResp)
		if err != nil {
			return false, "", err
		}

		if sealResp.StatusCode != 0 {
			fmt.Printf("%s\n", string(returnBody))
			return false, "", nil
		}

		return true, sealResp.Path, nil
	} else {
		resp.Body.Close()
		return false, "", fmt.Errorf("Seal error code is: %d", resp.StatusCode)
	}

}

func (sw *SWorker) unseal(path string) ([]byte, error) {
	// Generate request
	url := fmt.Sprintf("%s/storage/unseal", sw.url)
	body := fmt.Sprintf("{\"path\":\"%s\"}", path)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	// Request
	resp, err := sw.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Deal response
	if resp.StatusCode == 200 {
		returnBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return returnBody, nil
	} else {
		resp.Body.Close()
		return nil, fmt.Errorf("Unseal error code is: %d", resp.StatusCode)
	}
}

// TODO: config
var sw *SWorker = nil

func init() {
	sw = NewSWorker("http://127.0.0.1:12222/api/v0")
}
