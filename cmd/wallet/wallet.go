package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/util/file"
	"gopkg.in/sirupsen/logrus.v1"
	"gopkg.in/urfave/cli.v1"

	"github.com/kittycash/wallet/src/http"
	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/iko/transaction"
	"github.com/kittycash/wallet/src/util"
	"github.com/kittycash/wallet/src/wallet"
	"github.com/kittycash/wallet/src/dummy"
)

const (
	// TODO: Define proper values for these!
	TrustedGenPK     = "03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753"

	DefaultHttpAddress = "127.0.0.1:7908"

	DirRoot         = ".kittycash"
	DirChildWallets = "wallets"
)

const (
	fWalletDir = "wallet-dir"

	fHttpAddress = "http-address"
	fGUI         = "gui"
	fGUIDir      = "gui-dir"
	fTLS         = "tls"
	fTLSCert     = "tls-cert"
	fTLSKey      = "tls-key"

	fTest          = "test"
	fTestGenPK     = "test-gen-pk"
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
	homeDir   = file.UserHome()
	staticDir = func() string {
		if goPath := os.Getenv("GOPATH"); goPath != "" {
			return filepath.Join(goPath, "src/github.com/kittycash/wallet/wallet/dist")
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
			<<< TEST MODE >>>
		*/
		cli.BoolFlag{
			Name:  Flag(fTest),
			Usage: "whether to run wallet in test mode",
		},
		cli.StringFlag{
			Name:  Flag(fTestGenPK),
			Usage: "test mode trusted gen tx public key",
		},
	}
	app.Action = cli.ActionFunc(action)
}

func action(ctx *cli.Context) error {
	quit := util.CatchInterrupt()

	var (
		genPK     = cipher.MustPubKeyFromHex(TrustedGenPK)

		walletDir = ctx.String(fWalletDir)

		httpAddress = ctx.String(fHttpAddress)
		gui         = ctx.BoolT(fGUI)
		guiDir      = ctx.String(fGUIDir)
		tls         = ctx.Bool(fTLS)
		tlsCert     = ctx.String(fTLSCert)
		tlsKey      = ctx.String(fTLSKey)

		test = ctx.Bool(fTest)
	)

	// Test mode changes.
	if test {
		genPK = cipher.MustPubKeyFromHex(ctx.String(fTestGenPK))

		tempDir, err := ioutil.TempDir(os.TempDir(), "kc_wallet")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		walletDir = tempDir
	}

	// Prepare StateDB.
	stateDB := iko.NewMemoryState()

	// Prepare ChainDB.
	chainDB := new(dummy.Dummy)

	// Prepare blockchain config.
	bcConfig := &iko.BlockChainConfig{
		GenerationPK: genPK,
		TxAction: func(tx *transaction.Transaction) error {
			return nil
		},
	}

	// Prepare blockchain.
	bc, err := iko.NewBlockChain(bcConfig, chainDB, stateDB)
	if err != nil {
		return err
	}
	defer bc.Close()

	log.Info("finished preparing blockchain")

	// Prepare wallet.
	if err := wallet.SetRootDir(walletDir); err != nil {
		return err
	}
	walletManager, err := wallet.NewManager()
	if err != nil {
		return err
	}

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
			IKO:    bc,
			Wallet: walletManager,
			Conn:   new(dummy.Dummy),
		},
	)
	if err != nil {
		return err
	}
	defer httpServer.Close()

	<-quit
	return nil
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}
