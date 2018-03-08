package main

import (
	"github.com/kittycash/wallet/src/http"
	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/rpc"
	"github.com/kittycash/wallet/src/wallet"
	"github.com/skycoin/skycoin/src/cipher"
	"gopkg.in/sirupsen/logrus.v1"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"os/signal"
)

const (
	Init       = "init"
	RootPubKey = "root-public-key"
	RootSecKey = "root-secret-key"
	RootNonce  = 885560
	TxPubKey   = "tx-public-key"

	TestMode     = "test"
	TestTxCount  = "test-tx-count"
	TestTxSecKey = "test-tx-secret-key"

	CXODir             = "cxo-dir"
	CXOAddress         = "cxo-address"
	CXORPCAddress      = "cxo-rpc-address"
	DiscoveryAddresses = "messenger-addresses"

	WalletDir = "wallet-dir"

	HttpAddress = "http-address"
	GUI         = "gui"
	GUIDir      = "gui-dir"
	TLS         = "tls"
	TLSCert     = "tls-cert"
	TLSKey      = "tls-key"

	RPCAddress  = "rpc-address"
	RemoteClose = "remote-close"
)

func Flag(flag string, short ...string) string {
	if len(short) == 0 {
		return flag
	}
	return flag + ", " + short[0]
}

var (
	app = cli.NewApp()
	log = logrus.New()
)

func init() {
	app.Name = "iko"
	app.Description = "kittycash initial coin offering service"
	app.Flags = cli.FlagsByName{
		/*
			<<< MASTER >>>
		*/
		cli.StringFlag{
			Name:  Flag(RootPubKey, "rpk"),
			Usage: "public key to use as main blockchain signer",
		},
		cli.StringFlag{
			Name:  Flag(RootSecKey, "rsk"),
			Usage: "secret key to use as main blockchain signer",
		},
		cli.StringFlag{
			Name:  Flag(TxPubKey, "tpk"),
			Usage: "public key that is trusted for transactions",
		},
		cli.BoolFlag{
			Name:  Flag(Init),
			Usage: "whether to init the root if it doesn't exist",
		},
		/*
			<<< TEST MODE >>>
		*/
		cli.BoolFlag{
			Name:  Flag(TestMode, "t"),
			Usage: "whether to use test data for run",
		},
		cli.IntFlag{
			Name:  Flag(TestTxCount, "tc"),
			Usage: "only valid in test mode, injects a number of initial transactions for testing",
		},
		cli.StringFlag{
			Name:  Flag(TestTxSecKey, "tsk"),
			Usage: "secret key for signing test transactions",
			Value: new(cipher.SecKey).Hex(),
		},
		/*
			<<< WALLET CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(WalletDir),
			Usage: "directory to store wallet files",
			Value: "./kc/wallet",
		},
		/*
			<<< CXO CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(CXODir),
			Usage: "directory to store cxo files",
			Value: "./kc/cxo",
		},
		cli.StringFlag{
			Name:  Flag(CXOAddress),
			Usage: "address to use to serve CXO",
			Value: "[::]:8123", // TODO: Determine a default value.
		},
		cli.StringSliceFlag{
			Name:  Flag(DiscoveryAddresses),
			Usage: "discovery addresses",
		},
		cli.StringFlag{
			Name:  Flag(CXORPCAddress),
			Usage: "address for CXO RPC, leave blank to disable CXO RPC",
		},
		/*
			<<< HTTP SERVER >>>
		*/
		cli.StringFlag{
			Name:  Flag(HttpAddress),
			Usage: "address to serve http server on",
			Value: "127.0.0.1:8080",
		},
		cli.BoolTFlag{
			Name:  Flag(GUI),
			Usage: "whether to enable gui",
		},
		cli.StringFlag{
			Name:  Flag(GUIDir),
			Usage: "directory to serve GUI from",
			Value: "./kc/static",
		},
		cli.BoolFlag{
			Name:  Flag(TLS),
			Usage: "whether to enable tls",
		},
		cli.StringFlag{
			Name:  Flag(TLSCert),
			Usage: "tls certificate file path",
		},
		cli.StringFlag{
			Name:  Flag(TLSKey),
			Usage: "tls key file path",
		},
		/*
			<<< RPC SERVER >>>
		*/
		cli.StringFlag{
			Name:  Flag(RPCAddress),
			Usage: "address used to serve rpc, keep empty to not serve rpc",
			Value: "[::]:7123", // TODO: determine default value.
		},
		cli.BoolFlag{
			Name:  Flag(RemoteClose),
			Usage: "whether to enable remote close",
		},
	}
	app.Action = cli.ActionFunc(action)
}

func action(ctx *cli.Context) error {
	quit := CatchInterrupt()

	var (
		rootPK = cipher.MustPubKeyFromHex(ctx.String(RootPubKey))
		rootSK = cipher.MustSecKeyFromHex(ctx.String(RootSecKey))
		txPK   = cipher.MustPubKeyFromHex(ctx.String(TxPubKey))
		doInit = ctx.Bool(Init)

		testMode  = ctx.Bool(TestMode)
		testCount = ctx.Int(TestTxCount)
		testSK    = cipher.MustSecKeyFromHex(ctx.String(TestTxSecKey))

		walletDir = ctx.String(WalletDir)

		cxoDir             = ctx.String(CXODir)
		cxoAddress         = ctx.String(CXOAddress)
		cxoRPCAddress      = ctx.String(CXORPCAddress)
		discoveryAddresses = ctx.StringSlice(DiscoveryAddresses)

		httpAddress = ctx.String(HttpAddress)
		gui         = ctx.BoolT(GUI)
		guiDir      = ctx.String(GUIDir)
		tls         = ctx.Bool(TLS)
		tlsCert     = ctx.String(TLSCert)
		tlsKey      = ctx.String(TLSKey)

		rpcAddress = ctx.String(RPCAddress)
		remoteClose = ctx.Bool(RemoteClose)
	)

	var (
		e        error
		stateDB  iko.StateDB
		cxoChain *iko.CXOChain
	)

	// Prepare StateDB.
	stateDB = iko.NewMemoryState()

	// Prepare ChainDB.
	cxoChain, e = iko.NewCXOChain(&iko.CXOChainConfig{
		Dir:                cxoDir,
		Public:             true,
		Memory:             testMode,
		MessengerAddresses: discoveryAddresses,
		CXOAddress:         cxoAddress,
		CXORPCAddress:      cxoRPCAddress,
		MasterRooter:       true,
		MasterRootPK:       rootPK,
		MasterRootSK:       rootSK,
		MasterRootNonce:    RootNonce,
	})
	if e != nil {
		return e
	}
	defer cxoChain.Close()

	// Prepare blockchain config.
	bcConfig := &iko.BlockChainConfig{
		GenerationPK: txPK,
		TxAction: func(tx *iko.Transaction) error {
			return nil
		},
	}

	// Prepare blockchain.
	bc, e := iko.NewBlockChain(bcConfig, cxoChain, stateDB)
	if e != nil {
		return e
	}
	defer bc.Close()

	if cxoChain != nil {
		cxoChain.RunTxService(iko.MakeTxChecker(bc))
	}

	if doInit || testMode {
		if e := cxoChain.MasterInitChain(); e != nil {
			return e
		}
	}

	log.Info("finished preparing blockchain")

	// Prepare test data.
	if testMode {
		var tx *iko.Transaction
		for i := 0; i < testCount; i++ {
			tx = iko.NewGenTx(iko.KittyID(i), testSK)

			log.WithField("tx", tx.String()).
				Debugf("test:tx_inject(%d)", i)

			if _, e := bc.InjectTx(tx); e != nil {
				return e
			}
		}
	}

	// Prepare wallet.
	if testMode {
		tempDir, e := ioutil.TempDir(os.TempDir(), "kc")
		if e != nil {
			return e
		}
		defer os.RemoveAll(tempDir)
		walletDir = tempDir
	}
	if e := wallet.SetRootDir(walletDir); e != nil {
		return e
	}
	walletManager, e := wallet.NewManager()
	if e != nil {
		return e
	}

	// Prepare http server.
	httpServer, e := http.NewServer(
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
		},
	)
	if e != nil {
		return e
	}
	defer httpServer.Close()

	// Prepare rpc server.
	rpcServer, e := rpc.NewServer(
		&rpc.ServerConfig{
			Address: rpcAddress,
			EnableRemoteQuit: remoteClose,
		},
		&rpc.Gateway{
			IKO:      bc,
			QuitChan: quit,
		},
	)
	if e != nil {
		return e
	}
	defer rpcServer.Close()

	<-quit
	return nil
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}

// CatchInterrupt catches Ctrl+C behaviour.
func CatchInterrupt() chan int {
	quit := make(chan int)
	go func(q chan<- int) {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		signal.Stop(sigChan)
		q <- 1
	}(quit)
	return quit
}
