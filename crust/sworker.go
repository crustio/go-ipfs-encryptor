package crust

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
)

var sealBlackSet map[cid.Cid]bool
var sealBlackList = []string{
	"QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc",
	"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn",
}

func init() {
	sealBlackSet = make(map[cid.Cid]bool)
	for _, v := range sealBlackList {
		c, _ := cid.Decode(v)
		sealBlackSet[c] = true
	}
}

type sealResponse struct {
	Path       string `json:"path"`
	Message    string `json:"message"`
	StatusCode int64  `json:"status_code"`
}

type SWorker struct {
	lock   sync.Mutex
	url    string
	client http.Client
}

func NewSWorker(url string) *SWorker {
	client := http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: false,
		},
	}
	return &SWorker{url: url, client: client}
}

func (sw *SWorker) SetUrl(url string) {
	sw.lock.Lock()
	defer sw.lock.Unlock()
	sw.url = url
}

func (sw *SWorker) GetUrl() string {
	sw.lock.Lock()
	defer sw.lock.Unlock()
	url := sw.url
	return url
}

func (sw *SWorker) StartSeal(ci cid.Cid) (bool, error) {
	// Not config sworker
	if len(sw.GetUrl()) == 0 {
		return false, nil
	}

	if _, ok := sealBlackSet[ci]; ok {
		return false, nil
	}

	// Generate request
	url := fmt.Sprintf("%s/storage/seal_start", sw.GetUrl())
	value := fmt.Sprintf("{\"cid\":\"%s\"}", ci.String())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(value)))
	if err != nil {
		return false, err
	}

	// Request
	resp, err := sw.client.Do(req)
	if err != nil {
		return false, err
	}

	// Deal response
	if resp.StatusCode == 200 {
		returnBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return false, err
		}
		sealResp := &sealResponse{}
		err = json.Unmarshal(returnBody, sealResp)
		if err != nil {
			return false, err
		}

		if sealResp.StatusCode != 0 {
			fmt.Printf("%s\n", string(returnBody))
			return false, nil
		}

		return true, nil
	} else {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return false, fmt.Errorf("Start seal error code is: %d", resp.StatusCode)
	}

}

func (sw *SWorker) Seal(ci cid.Cid, newBlock bool, value []byte) (bool, string, error) {
	// Not config sworker
	if len(sw.GetUrl()) == 0 {
		return false, "", nil
	}

	// Generate request
	url := fmt.Sprintf("%s/storage/seal?cid=%s&new_block=%t", sw.GetUrl(), ci.String(), newBlock)
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
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return false, "", fmt.Errorf("Seal error code is: %d", resp.StatusCode)
	}

}

func (sw *SWorker) EndSeal(ci cid.Cid) (bool, error) {
	// Not config sworker
	if len(sw.GetUrl()) == 0 {
		return false, nil
	}

	// Generate request
	url := fmt.Sprintf("%s/storage/seal_end", sw.GetUrl())
	value := fmt.Sprintf("{\"cid\":\"%s\"}", ci.String())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(value)))
	if err != nil {
		return false, err
	}

	// Request
	resp, err := sw.client.Do(req)
	if err != nil {
		return false, err
	}

	// Deal response
	if resp.StatusCode == 200 {
		returnBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return false, err
		}
		sealResp := &sealResponse{}
		err = json.Unmarshal(returnBody, sealResp)
		if err != nil {
			return false, err
		}

		if sealResp.StatusCode != 0 {
			fmt.Printf("%s\n", string(returnBody))
			return false, nil
		}

		return true, nil
	} else {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return false, fmt.Errorf("Start seal error code is: %d", resp.StatusCode)
	}

}

func (sw *SWorker) unseal(path string) ([]byte, error) {
	// Not config sworker
	if len(sw.GetUrl()) == 0 {
		return nil, fmt.Errorf("Missing crust config")
	}

	// Generate request
	url := fmt.Sprintf("%s/storage/unseal", sw.GetUrl())
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
	} else if resp.StatusCode == 404 {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return nil, nil
	} else {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Unseal error code is: %d", resp.StatusCode)
	}
}

var Worker *SWorker = nil

func init() {
	Worker = NewSWorker("")
}
