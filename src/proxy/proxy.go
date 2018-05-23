package proxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Config struct {
	Domain string
	TLS    bool
}

func (c *Config) TransformURL(originalURL *url.URL) string {

	for _, s := range []string{"http://", "https://"} {
		c.Domain = strings.TrimPrefix(c.Domain, s)
	}

	out := path.Join(c.Domain, originalURL.EscapedPath()) +
		"?" + originalURL.Query().Encode()

	if c.TLS {
		return "https://" + out
	} else {
		return "http://" + out
	}
}

type Proxy struct {
	c    *Config
	http *http.Client
}

func New(c *Config) (*Proxy, error) {
	return &Proxy{
		c: c,
		http: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   time.Second * 10,
		},
	}, nil
}

func (p *Proxy) Call(req *http.Request) (*http.Response, error) {
	return call(p, req, nil)
}

func (p *Proxy) Redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, p.c.TransformURL(r.URL), http.StatusMovedPermanently)
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type Changer func(body []byte, header http.Header) ([]byte, error)

func call(p *Proxy, req *http.Request, change Changer) (*http.Response, error) {
	newURL, err := url.Parse(p.c.TransformURL(req.URL))
	if err != nil {
		return nil, err
	}
	resp, err := p.http.Get(newURL.String())
	if err != nil {
		return nil, err
	}
	// Only change response if Changer is defined and returned status is 200.
	if change != nil && resp.StatusCode == http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if data, err = change(data, resp.Header); err != nil {
			return nil, err
		}
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	}
	return resp, nil
}
