package tools

import (
	"net/url"
	"strings"
	"path"
)

func TransformURL(original *url.URL, newDomain string, useTLS bool) string {
	for _, s := range []string{"http://", "https://"} {
		newDomain = strings.TrimPrefix(newDomain, s)
	}

	out := path.Join(newDomain, original.EscapedPath()) +
		"?" + original.Query().Encode()

	if useTLS {
		return "https://" + out
	} else {
		return "http://" + out
	}
}