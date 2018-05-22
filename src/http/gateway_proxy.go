package http

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kittycash/wallet/src/proxy"
)

func proxyGateway(m *http.ServeMux, p *proxy.Proxy) error {
	Handle(m, "/v1/kitty_count", "GET", tunnel(p))
	Handle(m, "/v1/kitty/", "GET", tunnel(p))
	Handle(m, "/v1/kitties", "GET", tunnel(p))
	Handle(m, "/v1/image/", "GET", tunnel(p))
	Handle(m, "/v1/balance/", "GET", tunnel(p))
	Handle(m, "/v1/ping", "GET", tunnel(p))
	Handle(m, "/v1/last_transfer", "GET", tunnel(p))
	Handle(m, "/v1/transfer", "POST", tunnel(p))
	return nil
}

func tunnel(p *proxy.Proxy) HandlerFunc {
	return proxyHandler(func(req *http.Request) (*http.Response, error) {
		return p.Call(req)
	})
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type ProxyAction func(req *http.Request) (*http.Response, error)

func proxyHandler(action ProxyAction) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		resp, err := action(r)
		if err != nil {
			return sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)

		data, _ := ioutil.ReadAll(resp.Body)
		_, err = w.Write(data)
		return err
	}
}
