# Kittycash Wallet

Where one claims ownership of dem' kitties.

## Test wallet


**Start Wallet Node in Test Mode**
This is so that nothing gets written to disk.

```
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/wallet/wallet.go \
--test=true \
--test-gen-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--http-address="127.0.0.1:6148"
```

## Run Wallet

```
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/wallet/wallet.go \
--http-address="127.0.0.1:6148"
```