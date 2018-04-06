package http

import (
	"net/http"
	"github.com/kittycash/wallet/src/connectivity"
	"fmt"
)

func connGateway(m *http.ServeMux, g connectivity.Connectivity) error {
	Handle(m, "/api/conn/all_statuses", "GET", allStatuses())
	Handle(m, "/api/conn/status", "GET", status(g))
	Handle(m, "/api/conn/reconnect", "GET", reconnect(g))
	return nil
}

func allStatuses() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		return sendJson(w, http.StatusOK, connectivity.Statuses())
	}
}

func status(g connectivity.Connectivity) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		status, err := g.Status()
		if err != nil {
			return sendJson(w, http.StatusInternalServerError,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		return sendJson(w, http.StatusOK, status)
	}
}

func reconnect(g connectivity.Connectivity) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		return sendJson(w, http.StatusOK, g.Reconnect())
	}
}