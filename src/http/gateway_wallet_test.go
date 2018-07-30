package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kittycash/wallet/src/wallet"
)

var CTApplicationFormHeaders = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}

type ResponseChecker func(*testing.T, *http.Response)

func TestWalletGateway(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "KittyCashTestWallet")
	require.NoError(t, err, "Created temp wallet directory")

	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err, "Remove temp wallet directory")
	}()

	manager, err := wallet.NewManager(&wallet.ManagerConfig{
		RootDir: tempDir,
	})
	require.NoError(t, err, "Should be able to create a wallet manager")

	// Get an http server
	mux := http.NewServeMux()

	err = walletGateway(mux, manager)

	require.NoError(t, err,
		"Shouldn't have an error initializing the walletGateway")

	testCases := []struct {
		name          string
		endpoint      string
		body          string
		method        string
		headers       http.Header
		responseCode  int
		checkResponse ResponseChecker
	}{
		/* /v1/wallets/seed tests */
		{
			endpoint:      "/v1/wallets/seed",
			name:          "No seedBitSize provided",
			method:        http.MethodPost,
			headers:       CTApplicationFormHeaders,
			responseCode:  http.StatusOK,
			checkResponse: validSeedChecker,
		},
		{
			endpoint:      "/v1/wallets/seed",
			name:          "Non-default seedBitSize provided",
			method:        http.MethodPost,
			body:          "seedBitSize=256",
			headers:       CTApplicationFormHeaders,
			responseCode:  http.StatusOK,
			checkResponse: validSeedChecker,
		},
		{
			endpoint:      "/v1/wallets/seed",
			name:          "Invalid seedBitSize provided",
			method:        http.MethodPost,
			body:          "seedBitSize=23",
			headers:       CTApplicationFormHeaders,
			responseCode:  http.StatusBadRequest,
			checkResponse: alwaysValidChecker,
		},
	}

	for _, testCase := range testCases {
		fullName := fmt.Sprintf("%s %s - %d - %s",
			testCase.method, testCase.endpoint, testCase.responseCode, testCase.name)

		t.Run(fullName, func(t *testing.T) {

			requestBody := bytes.NewBufferString(testCase.body)
			request := httptest.NewRequest(testCase.method, testCase.endpoint, requestBody)
			require.NotNil(t, request, "Should create the test request")
			request.Header = testCase.headers

			responseRecorder := httptest.NewRecorder()
			require.NotNil(t, responseRecorder, "Should create responseRecorder")

			mux.ServeHTTP(responseRecorder, request)

			response := responseRecorder.Result()
			require.NotNil(t, response, "Should return the response")

			require.Equal(t, testCase.responseCode, response.StatusCode,
				"Should return expected status code")

			testCase.checkResponse(t, response)
		})
	}
}

func validSeedChecker(t *testing.T, response *http.Response) {
	decoder := json.NewDecoder(response.Body)
	require.NotNil(t, decoder, "Should be able to create a JSON decoder")

	var seedReply SeedReply
	err := decoder.Decode(&seedReply)
	require.NoError(t, err, "Should be able to decode a SeedReply")
	require.NotEmpty(t, seedReply.Seed, "Should have a non-empty seed field")
}

func alwaysValidChecker(t *testing.T, response *http.Response) {
}
