package http

import (
	"net/http"
)

func marketKitties(m *http.ServeMux) error {
	Handle(m, "/api/count", http.MethodGet, count())
	Handle(m, "/api/entry", http.MethodGet, entry())
	Handle(m, "/api/entries", http.MethodGet, entries())
	return nil
}

func count() HandlerFunc {
	return dummyMarketHandler()
}

func entry() HandlerFunc {
	return dummyMarketHandler()
}

func entries() HandlerFunc {
	return dummyMarketHandler()
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type MHAction func(req *http.Request) (*http.Response, error)

//func marketHandler(action MHAction) HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
//		resp, err := action(r)
//		if err != nil {
//			return sendJson(w, http.StatusBadRequest,
//				fmt.Sprintf("Error: %s", err.Error()))
//		}
//		data, _ := ioutil.ReadAll(resp.Body)
//
//		w.Header().Set("Content-Type", "application/json")
//		w.WriteHeader(resp.StatusCode)
//		_, err = w.Write(data)
//		return err
//	}
//}

func dummyMarketHandler() HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request, _ *Path) error {
		w.WriteHeader(http.StatusNotImplemented)
		return nil
	}
}