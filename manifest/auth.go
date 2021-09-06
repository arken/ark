package manifest

import (
	"errors"
	"net/url"

	"github.com/arken/ark/manifest/upstream"
)

func Auth(path string) (result upstream.Guard, err error) {
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	upstream, ok := upstream.AvailableUpstreams[url.Host]
	if !ok {
		return nil, errors.New("unknown upstream")
	}
	return upstream.Auth(path)
}
