package http

import (
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
	Handle(m, "/v1/traits", "GET", tunnel(p))
	Handle(m, "/v1/trait_image/", "GET", tunnel(p))
	Handle(m, "/v1/redeem", "POST", tunnel(p))
	Handle(m, "/v1/scoreboard/", "GET", tunnel(p))
	return nil
}

func tunnel(p *proxy.Proxy) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ *Path) error {
		p.Redirect(w, r)
		return nil
	}
}
