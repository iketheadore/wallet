package kitties

import (
	"encoding/json"
	"github.com/kittycash/kitty-api/src/api"
	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type ManagerConfig struct {
	KittyAPIDomain string
}

func (mc *ManagerConfig) TransformURL(originalURL *url.URL) string {
	return path.Join(append(
		[]string{mc.KittyAPIDomain},
		originalURL.EscapedPath())...)
}

type Manager struct {
	c    *ManagerConfig
	iko  *iko.BlockChain
	http *http.Client
}

func (m *Manager) Count(req *http.Request) (*http.Response, error) {
	return m.do(req, func(resp *http.Response) (*http.Response, error) {
		return resp, nil
	})
}

func (m *Manager) Entry(req *http.Request) (*http.Response, error) {
	return m.do(req, func(resp *http.Response) (*http.Response, error) {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var (
			out = new(api.EntryOut)
		)
		if err := json.Unmarshal(data, out); err != nil {
			return nil, errRespCorrupt(err)
		}
		// TODO: Complete!
		return resp, nil
	})
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type ResponseChanger func(resp *http.Response) (*http.Response, error)

func (m *Manager) do(req *http.Request, changer ResponseChanger) (*http.Response, error) {
	var (
		err  error
		resp *http.Response
	)
	req.URL, err = url.Parse(m.c.TransformURL(req.URL))
	if err != nil {
		return nil, err
	}
	resp, err = m.http.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return changer(resp)
	default:
		return resp, nil
	}
}

func processResp(resp *http.Response) ([]byte, error) {
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errNot200(raw)
	}
	return raw, nil
}

func errRespCorrupt(err error) error {
	return errors.WithMessage(err,
		"response data is corrupt")
}

func errNot200(raw []byte) error {
	return errors.Errorf(
		"http status is not 200: %s", string(raw))
}

func errURLTransFail(err error) error {
	return errors.WithMessage(err,
		"failed to transform URL")
}
