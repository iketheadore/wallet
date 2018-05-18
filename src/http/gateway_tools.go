package http

import (
	"net/http"
	"github.com/kittycash/kittiverse/src/kitty"
	"strconv"
	"fmt"
	"github.com/kittycash/wallet/src/tools"
)

func toolsGateway(m *http.ServeMux) error {
	Handle(m, "/api/tools/sign_transfer_params", "GET", signTransferParams())
	return nil
}

func signTransferParams() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, err := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vKittyID         = r.PostFormValue("kittyID")
					vLastTransferSig = r.PostFormValue("lastTransferSig")
					vToAddress       = r.PostFormValue("toAddress")
					vSecretKey       = r.PostFormValue("secretKey")
				)

				var kittyID kitty.ID
				{
					v, err := strconv.ParseUint(vKittyID, 10, 64)
					if err != nil {
						return false, sendJson(w, http.StatusBadRequest,
							fmt.Sprintf("Error: %s", err))
					}
					kittyID = kitty.ID(v)
				}

				tools.SignTransferParams()
			},
		})

		return nil
	}
}