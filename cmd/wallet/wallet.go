package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/util/file"
	"gopkg.in/urfave/cli.v1"

	"github.com/watercompany/kittycash-wallet/src/http"
	"github.com/watercompany/kittycash-wallet/src/proxy"
	"github.com/watercompany/kittycash-wallet/src/util"
	"github.com/watercompany/kittycash-wallet/src/wallet"
)

//Production override variables
const (
	DirChildWalletsProd     = "wallets"
	DefaultProxyAddressProd = "api.kittycash.io"
)

const (
	DefaultHttpAddress  = "127.0.0.1:7908"
	DefaultProxyAddress = "staging-api.kittycash.io"
	DirRoot             = ".kittycash"
	DirChildWallets     = "staging-wallets"
)

const (
	fWalletDir = "wallet-dir"

	fProxyDomain = "proxy-domain"
	fProxyTLS    = "proxy-tls"

	fHttpAddress = "http-address"
	fGUI         = "gui"
	fGUIDir      = "gui-dir"
	fTLS         = "tls"
	fTLSCert     = "tls-cert"
	fTLSKey      = "tls-key"
	fProduction  = "production"

	fTest = "test"
)

func Flag(flag string, short ...string) string {
	if len(short) == 0 {
		return flag
	}
	return flag + ", " + short[0]
}

var (
	app = cli.NewApp()
	log = &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	homeDir = file.UserHome()

	staticDir = func() string {
		if goPath := os.Getenv("GOPATH"); goPath != "" {
			return filepath.Join(goPath, "src/github.com/watercompany/kittycash-wallet/wallet/dist")
		}
		return "./static/dist"
	}()
)

func init() {
	app.Name = "wallet"
	app.Description = "kitty cash wallet executable"
	app.Flags = cli.FlagsByName{
		/*
			<<< WALLET CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(fWalletDir),
			Usage: "directory to store wallet files",
			Value: filepath.Join(homeDir, DirRoot, DirChildWallets),
		},
		/*
			<<< PROXY CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(fProxyDomain),
			Usage: "domain to proxy kitty-api requests to",
			Value: DefaultProxyAddress,
		},
		cli.BoolTFlag{
			Name:  Flag(fProxyTLS),
			Usage: "whether to use TLS to communicate to kitty-api domain",
		},
		/*
			<<< HTTP SERVER >>>
		*/
		cli.StringFlag{
			Name:  Flag(fHttpAddress),
			Usage: "address to serve http server on",
			Value: DefaultHttpAddress,
		},
		cli.BoolTFlag{
			Name:  Flag(fGUI),
			Usage: "whether to enable gui",
		},
		cli.StringFlag{
			Name:  Flag(fGUIDir),
			Usage: "directory to serve GUI from",
			Value: staticDir,
		},
		cli.BoolFlag{
			Name:  Flag(fTLS),
			Usage: "whether to enable tls",
		},
		cli.StringFlag{
			Name:  Flag(fTLSCert),
			Usage: "tls certificate file path",
		},
		cli.StringFlag{
			Name:  Flag(fTLSKey),
			Usage: "tls key file path",
		},
		/*
			<<< PRODUCTION / STAGING >>>
		*/
		cli.BoolFlag{
			Name:  Flag(fProduction),
			Usage: "whether to enable production",
		},
		/*
			<<< TEST MODE >>>
		*/
		cli.BoolFlag{
			Name:  Flag(fTest),
			Usage: "whether to run wallet in test mode",
		},
	}
	app.Action = cli.ActionFunc(action)
}

func action(ctx *cli.Context) error {
	quit := util.CatchInterrupt()

	var (
		walletDir = ctx.String(fWalletDir)

		proxyDomain = ctx.String(fProxyDomain)
		proxyTLS    = ctx.BoolT(fProxyTLS)

		httpAddress = ctx.String(fHttpAddress)
		gui         = ctx.BoolT(fGUI)
		guiDir      = ctx.String(fGUIDir)
		tls         = ctx.Bool(fTLS)
		tlsCert     = ctx.String(fTLSCert)
		tlsKey      = ctx.String(fTLSKey)
		production  = ctx.Bool(fProduction)

		test = ctx.Bool(fTest)
	)

	// If we are running in production, override all the other variables
	if production {
		log.Printf("Wallet is running in production")
		walletDir = filepath.Join(homeDir, DirRoot, DirChildWalletsProd)
		proxyDomain = DefaultProxyAddressProd
	} else {
		log.Printf("Wallet is running in staging")
	}

	// Test mode changes.
	if test {
		tempDir, err := ioutil.TempDir(os.TempDir(), "kc_wallet")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		walletDir = tempDir
	}

	// Prepare wallet.
	walletManager, err := wallet.NewManager(&wallet.ManagerConfig{
		RootDir: walletDir,
	})
	if err != nil {
		return err
	}
	log.Printf("INIT: wallet directory is '%s' (TEST:%v).",
		walletDir, test)

	// Prepare proxy.
	proxyManager, err := proxy.New(&proxy.Config{
		Domain: proxyDomain,
		TLS:    proxyTLS,
	})
	if err != nil {
		return err
	}
	log.Printf("INIT: proxy is relaying requests to '%s' (TLS:%v).",
		proxyDomain, proxyTLS)

	// Prepare http server.
	httpServer, err := http.NewServer(
		&http.ServerConfig{
			Address:     httpAddress,
			EnableGUI:   gui,
			GUIDir:      guiDir,
			EnableTLS:   tls,
			TLSCertFile: tlsCert,
			TLSKeyFile:  tlsKey,
		},
		&http.Gateway{
			Wallet: walletManager,
			Proxy:  proxyManager,
		},
	)
	if err != nil {
		return err
	}
	defer httpServer.Close()
	log.Printf("INIT: http server is serving on '%s' (TLS:%v).",
		httpAddress, tls)

	<-quit
	return nil
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}
