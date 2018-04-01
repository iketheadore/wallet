package kitties

import (
	"bytes"
	"encoding/json"
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
	return m.do(req, func(body []byte) ([]byte, error) {
		return body, nil
	})
}

type EntryOut struct {
	Entry *iko.KittyEntry `json:"entry"`
}

func (m *Manager) Entry(req *http.Request) (*http.Response, error) {
	return m.do(req, func(body []byte) ([]byte, error) {
		var (
			out = new(EntryOut)
		)
		if err := json.Unmarshal(body, out); err != nil {
			return nil, errRespCorrupt(err)
		}
		state, ok := m.iko.GetKittyState(out.Entry.ID)
		if !ok {
			return nil, errNoStateInfo(out.Entry.ID)
		}
		out.Entry.Address = state.Address.String()
		body, _ = json.Marshal(out)
		return body, nil
	})
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type BodyChanger func(body []byte) ([]byte, error)

func (m *Manager) do(req *http.Request, changer BodyChanger) (*http.Response, error) {
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
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		data, err = changer(data)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	}
	return resp, nil
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

func errNoStateInfo(kittyID iko.KittyID) error {
	return errors.Errorf(
		"no state information for kitty of ID '%d'", kittyID)
}
