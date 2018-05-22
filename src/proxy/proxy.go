package proxy

import (
	"net/url"
	"path"
	"strings"
)

type Config struct {
	Domain string
	TLS    bool
}

func (c *Config) TransformURL(originalURL *url.URL) string {

	for _, s := range []string{"http://", "https://"} {
		c.Domain = strings.TrimPrefix(c.Domain, s)
	}

	out := path.Join(c.Domain, originalURL.EscapedPath())+
		"?" + originalURL.Query().Encode()

	if c.TLS {
		return "https://" + out
	} else {
		return "http://" + out
	}
}