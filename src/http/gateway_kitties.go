package http

import (
	"github.com/kittycash/wallet/src/kitties"
	"net/http"
	"fmt"
	"bytes"
	"github.com/kittycash/wallet/src/iko"
)

func marketKitties(m *http.ServeMux, g *kitties.Manager, bc *iko.BlockChain) error {
	Handle(m, "/api/count", http.MethodGet, count(g))
	Handle(m, "/api/entry", http.MethodGet, entry(g, bc))
	Handle(m, "/api/entries", http.MethodGet, entries(g, bc))
	return nil
}

func count(g *kitties.Manager) HandlerFunc {
	return marketHandler(func(req *http.Request) (*http.Response, error) {
		return g.Count(req)
	})
}

func entry(g *kitties.Manager, bc *iko.BlockChain) HandlerFunc {
	return marketHandler(func(req *http.Request) (*http.Response, error) {
		return g.Entry(bc, req)
	})
}

func entries(g *kitties.Manager, bc *iko.BlockChain) HandlerFunc {
	return marketHandler(func(req *http.Request) (*http.Response, error) {
		return g.Entries(bc, req)
	})
}



/*
	<<< HELPER FUNCTIONS >>>
*/

type MHAction func(req *http.Request) (*http.Response, error)

func marketHandler(action MHAction) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		resp, err := action(r)
		if err != nil {
			sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		b := bytes.NewBuffer(make([]byte, resp.ContentLength))
		if err := resp.Write(b); err != nil {
			sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		w.WriteHeader(resp.StatusCode)
		_, err = w.Write(b.Bytes())
		return err
	}
}