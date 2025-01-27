package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/watercompany/kittycash-wallet/src/wallet"
)

func walletGateway(m *http.ServeMux, g *wallet.Manager) error {
	Handle(m, "/v1/wallets/refresh", "GET", refreshWallets(g))
	Handle(m, "/v1/wallets/list", "GET", listWallets(g))
	Handle(m, "/v1/wallets/new", "POST", newWallet(g))
	Handle(m, "/v1/wallets/delete", "POST", deleteWallet(g))
	Handle(m, "/v1/wallets/get", "POST", getWallet(g))
	Handle(m, "/v1/wallets/get_paginated", "POST", getWalletPaginated(g))
	Handle(m, "/v1/wallets/rename", "POST", renameWallet(g))
	Handle(m, "/v1/wallets/seed", "POST", newSeed())
	return nil
}

func refreshWallets(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		// Send json response with 500 status code if error.
		if e := g.Refresh(); e != nil {
			return sendJson(w, http.StatusInternalServerError,
				fmt.Sprintf("message: '%s'", e))
		}
		// Send json response with 200 status code if error is nil.
		return sendJson(w, http.StatusOK, true)
	}
}

type WalletsReply struct {
	Wallets []wallet.Stat `json:"wallets"`
}

func listWallets(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		return sendJson(w, http.StatusOK, WalletsReply{
			Wallets: g.ListWallets(),
		})
	}
}

func newWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel     = r.PostFormValue("label")
					vSeed      = r.PostFormValue("seed")
					vAddresses = r.PostFormValue("aCount")
					vEncrypted = r.PostFormValue("encrypted")
					vPassword  = r.PostFormValue("password")
				)

				encrypted, e := strconv.ParseBool(vEncrypted)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e))
				}

				// Options to pass to g.NewWallet()
				opts := wallet.Options{
					Label:     vLabel,
					Seed:      vSeed,
					Encrypted: encrypted,
					Password:  vPassword,
				}

				/**
				 * Verify that all values are correct
				 * Respond if options are not correct
				 */

				if e := opts.Verify(); e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				// Get aCount and convert it to int.
				aCount, e := strconv.Atoi(vAddresses)
				if e != nil {
					return false, sendJson(w, http.StatusNotAcceptable,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				if e := g.NewWallet(&opts, aCount); e != nil {
					return false, sendJson(w, http.StatusInternalServerError,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				return true, sendJson(w, http.StatusOK, true)
			},
		})
		return e
	}
}

func deleteWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel = r.PostFormValue("label")
				)
				if r.Body == nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprint("request body missing"))
				}
				if e := g.DeleteWallet(vLabel); e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: failed to delete wallet of label '%s': %v",
							vLabel, e))
				}
				return true, sendJson(w, http.StatusOK, true)
			},
		})
		return e
	}
}

func getWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel     = r.PostFormValue("label")
					vPassword  = r.PostFormValue("password") // Optional.
					vAddresses = r.PostFormValue("aCount")   // Optional.
				)
				if r.Body == nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprint("request body missing"))
				}
				var addresses int
				if vAddresses != "" {
					var e error
					addresses, e = strconv.Atoi(vAddresses)
					if e != nil {
						return false, sendJson(w, http.StatusBadRequest,
							fmt.Sprintf("Error: %s", e))
					}
				}
				fw, e := g.DisplayWallet(vLabel, vPassword, addresses)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %v", e))
				} else if fw == nil {
					// WORKAROUND: happens only on panic within DisplayWallet function.
					// 		panic should only happen when wrong password is given.
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %v", wallet.ErrInvalidPassword))
				}
				return true, sendJson(w, http.StatusOK, fw)
			},
		})
		return e
	}
}

func getWalletPaginated(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, err := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel      = r.PostFormValue("label")
					vPassword   = r.PostFormValue("password") // Optional.
					vStartIndex = r.PostFormValue("startIndex")
					vPageSize   = r.PostFormValue("pageSize")
					vForceTotal = r.PostFormValue("forceTotal")
				)
				if r.Body == nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprint("request body missing"))
				}
				var startIndex int
				if vStartIndex != "" {
					var err error
					if startIndex, err = strconv.Atoi(vStartIndex); err != nil {
						return false, sendJson(w, http.StatusBadRequest,
							fmt.Sprintf("invalid startIndex: %s", err.Error()))
					}
				}
				var pageSize int
				if vPageSize != "" {
					var err error
					if pageSize, err = strconv.Atoi(vPageSize); err != nil {
						return false, sendJson(w, http.StatusBadRequest,
							fmt.Sprintf("invalid pageSize: %s", err.Error()))
					}
				}
				var forceTotal int
				if vForceTotal != "" {
					var err error
					if forceTotal, err = strconv.Atoi(vForceTotal); err != nil {
						return false, sendJson(w, http.StatusBadRequest,
							fmt.Sprintf("invalid forceTotal: %s", err.Error()))
					}
				} else {
					forceTotal = -1
				}

				fw, err := g.DisplayPaginatedWallet(vLabel, vPassword, startIndex, pageSize, forceTotal)
				if err != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %v", err))
				} else if fw == nil {
					// WORKAROUND: happens only on panic within DisplayWallet function.
					// 		panic should only happen when wrong password is given.
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %v", wallet.ErrInvalidPassword))
				}
				return true, sendJson(w, http.StatusOK, fw)
			},
		})
		return err
	}
}

func renameWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel    = r.PostFormValue("label")
					vNewLabel = r.PostFormValue("newLabel")
				)

				if e := g.RenameWallet(vLabel, vNewLabel); e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				return true, sendJson(w, http.StatusOK, true)
			},
		})
		return e
	}
}

type SeedReply struct {
	Seed string `json:"seed"`
}

func newSeed() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				vSeedBitSize := r.PostFormValue("seedBitSize")
				seedBitSize, e := wallet.SeedBitSizeFromString(vSeedBitSize)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e))
				}
				seed, e := wallet.NewSeed(seedBitSize)
				if e != nil {
					return false, sendJson(w, http.StatusInternalServerError,
						fmt.Sprintf("Error: %v", e))
				}
				return true, sendJson(w, http.StatusOK, SeedReply{
					Seed: seed,
				})
			},
		})
		return e
	}
}
