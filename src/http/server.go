package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/skycoin/skycoin/src/util/iputil"
	"gopkg.in/sirupsen/logrus.v1"
)

const (
	indexFileName = "index.html"
)

type ServerConfig struct {
	Address        string
	KittyAPIDomain string
	EnableGUI      bool
	GUIDir         string
	EnableTLS      bool
	TLSCertFile    string
	TLSKeyFile     string
}

type SplitAddressOut struct {
	Address   string
	Port      uint16
	Localhost bool
}

func (sc *ServerConfig) SplitAddress() (*SplitAddressOut, error) {
	var (
		addr = sc.Address
		port = uint16(0)
	)
	if strings.Contains(sc.Address, ":") {
		var err error
		addr, port, err = iputil.SplitAddr(sc.Address)
		if err != nil {
			return nil, err
		}
	}

	localhost := iputil.IsLocalhost(addr)

	if localhost && port == 0 {
		return nil, errors.New("localhost with no port specified is unsupported")
	}

	return &SplitAddressOut{
		Address:   addr,
		Port:      port,
		Localhost: localhost,
	}, nil
}

type Server struct {
	c    *ServerConfig
	srv  *http.Server
	mux  *http.ServeMux
	api  *Gateway
	quit chan struct{}
}

func NewServer(config *ServerConfig, api *Gateway) (*Server, error) {
	var server = &Server{
		c:    config,
		mux:  http.NewServeMux(),
		api:  api,
		quit: make(chan struct{}),
	}
	if e := server.prepareMux(); e != nil {
		return nil, e
	}
	a, err := config.SplitAddress()
	if err != nil {
		return nil, errors.WithMessage(err, "provided address not supported")
	}
	go server.serve(a)
	return server, nil
}

func (s *Server) serve(a *SplitAddressOut) {
	s.srv = &http.Server{
		Addr:    s.c.Address,
		Handler: HostCheck(logrus.New(), a, s.mux),
	}
	if s.c.EnableTLS {
		for {
			if e := s.srv.ListenAndServeTLS(s.c.TLSCertFile, s.c.TLSKeyFile); e != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				break
			}
		}
	} else {
		for {
			if e := s.srv.ListenAndServe(); e != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				break
			}
		}
	}
	s.srv = nil
}

func (s *Server) prepareMux() error {
	if s.c.EnableGUI {
		if e := s.prepareGUI(); e != nil {
			return e
		}
	}
	return s.api.host(s.mux)
}

func (s *Server) prepareGUI() error {
	appLoc := s.c.GUIDir
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := path.Join(appLoc, indexFileName)
		http.ServeFile(w, r, page)
	})

	list, _ := ioutil.ReadDir(appLoc)
	for _, fInfo := range list {
		route := fmt.Sprintf("/%s", fInfo.Name())
		if fInfo.IsDir() {
			route += "/"
		}
		s.mux.Handle(route, http.FileServer(http.Dir(appLoc)))
	}
	return nil
}

// Close quits the http server.
func (s *Server) Close() {
	if s.quit != nil {
		close(s.quit)
		if s.srv != nil {
			s.srv.Close()
		}
	}
}

func HostCheck(log *logrus.Logger, a *SplitAddressOut, mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "" && a.Localhost &&
			r.Host != fmt.Sprintf("127.0.0.1:%d", a.Port) &&
			r.Host != fmt.Sprintf("localhost:%d", a.Port) {
			err := fmt.Sprintf("Detected DNS rebind attempt - configured-host=%s header-host=%s", r.Host, r.Host)
			log.Warn(err)
			http.Error(w, err, http.StatusForbidden)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
