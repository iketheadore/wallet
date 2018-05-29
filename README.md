# Kittycash Wallet

Where one claims ownership of dem' kitties.

```sh
$ wallet -h

NAME:
   wallet - A new cli application

USAGE:
   wallet [global options] command [command options] [arguments...]

VERSION:
   0.0.0

DESCRIPTION:
   kitty cash wallet executable

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --wallet-dir value    directory to store wallet files (default: "/Users/evanlinjin/.kittycash/wallets")
   --proxy-domain value  domain to proxy kitty-api requests to (default: "api.kittycash.com")
   --proxy-tls           whether to use TLS to communicate to kitty-api domain
   --http-address value  address to serve http server on (default: "127.0.0.1:7908")
   --gui                 whether to enable gui
   --gui-dir value       directory to serve GUI from (default: "/Users/evanlinjin/go/src/github.com/kittycash/wallet/wallet/dist")
   --tls                 whether to enable tls
   --tls-cert value      tls certificate file path
   --tls-key value       tls key file path
   --test                whether to run wallet in test mode
   --help, -h            show help
   --version, -v         print the version
```

## Run Wallet

**Run the wallet backend.**

```
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/wallet/wallet.go
```

**Run the wallet frontend.**

Refer to [/electron/README.md](/electron/README.md). 

## Test wallet

**Start wallet backend in test mode.**

This is so that nothing gets written to disk. We have also set proxy domain to `staging-api.kittycash.com` instead of `api.kittycash.com`.

```
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/wallet/wallet.go \
--test=true \
--test-gen-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--proxy-domain="staging-api.kittycash.com" \
--proxy-tls=true \
--http-address="127.0.0.1:6148"
```

## Endpoints documentation

Refer to the [Postman](https://www.getpostman.com) collection located at [/docs/Wallet.postman_collection.json](/docs/Wallet.postman_collection.json) .