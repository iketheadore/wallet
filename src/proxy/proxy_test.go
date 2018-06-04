package proxy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_TransformURL(t *testing.T) {
	cases := []struct {
		Config Config
		URL    string
		Exp    string
	}{
		{
			Config: Config{
				Domain: "api.google.com",
				TLS:    true,
			},
			URL: "http://127.0.0.1:1234/v1/kitty/0?size=large",
			Exp: "https://api.google.com/v1/kitty/0?size=large",
		},
	}

	for _, c := range cases {
		u, err := url.Parse(c.URL)
		require.NoError(t, err)
		require.Equal(t, c.Exp, c.Config.TransformURL(u))
	}
}
