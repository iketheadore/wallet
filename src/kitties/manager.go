package kitties

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pkg/errors"

	"github.com/kittycash/kittiverse/src/kitty"

	"github.com/kittycash/wallet/src/iko"
)

type ManagerConfig struct {
	KittyAPIDomain string
	TLS            bool
}

func (mc *ManagerConfig) TransformURL(originalURL *url.URL) string {
	out := path.Join(append(
		[]string{mc.KittyAPIDomain},
		originalURL.EscapedPath())...) + "?" + originalURL.Query().Encode()
	if mc.TLS {
		out = "https://" + out
	} else {
		out = "http://" + out
	}
	return out
}

type Manager struct {
	c    *ManagerConfig
	http *http.Client
}

func NewManager(c *ManagerConfig) (*Manager, error) {
	return &Manager{
		c: c,
		http: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   time.Second * 10,
		},
	}, nil
}

func (m *Manager) Count(req *http.Request) (*http.Response, error) {
	return m.do(req, func(body []byte) ([]byte, error) {
		return body, nil
	})
}

type EntryOut struct {
	Entry *kitty.ReadableKitty `json:"entry"`
}

func (m *Manager) Entry(bc *iko.BlockChain, req *http.Request) (*http.Response, error) {
	return m.do(req, func(body []byte) ([]byte, error) {
		var (
			out = new(EntryOut)
		)
		if err := json.Unmarshal(body, out); err != nil {
			return nil, errRespCorrupt(err)
		}
		state, ok := bc.GetKittyState(out.Entry.Info.ID)
		if !ok {
			return nil, errNoStateInfo(out.Entry.Info.ID)
		}
		out.Entry.Meta.Address = state.Address.String()
		body, _ = json.Marshal(out)
		return body, nil
	})
}

type EntriesOut struct {
	TotalCount int64                  `json:"total_count"`
	PageCount  int                    `json:"page_count"`
	Entries    []*kitty.ReadableKitty `json:"entries"`
}

func (m *Manager) Entries(bc *iko.BlockChain, req *http.Request) (*http.Response, error) {
	return m.do(req, func(body []byte) ([]byte, error) {
		var (
			out = new(EntriesOut)
		)
		if err := json.Unmarshal(body, out); err != nil {
			return nil, errRespCorrupt(err)
		}
		for i, entry := range out.Entries {
			state, ok := bc.GetKittyState(entry.Info.ID)
			if !ok {
				return nil, errors.WithMessage(errNoStateInfo(entry.Info.ID),
					fmt.Sprintf("failed at index %d", i))
			}
			out.Entries[i].Meta.Address = state.Address.String()
		}
		body, _ = json.Marshal(out)
		return body, nil
	})
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type BodyChanger func(body []byte) ([]byte, error)

func (m *Manager) do(req *http.Request, changer BodyChanger) (*http.Response, error) {
	newURL, err := url.Parse(m.c.TransformURL(req.URL))
	if err != nil {
		return nil, err
	}
	resp, err := m.http.Get(newURL.String())
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

func errNoStateInfo(kittyID kitty.ID) error {
	return errors.Errorf(
		"no state information for kitty of ID '%d'", kittyID)
}
