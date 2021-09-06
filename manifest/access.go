package manifest

import (
	"errors"
	"net/url"

	"github.com/arken/ark/manifest/upstream"
)

func (m *Manifest) HaveWriteAccess() (bool, error) {
	url, err := url.Parse(m.url)
	if err != nil {
		return false, err
	}

	upstream, ok := upstream.AvailableUpstreams[url.Host]
	if !ok {
		return true, errors.New("unknown upstream")
	}
	return upstream.HaveWriteAccess(m.gitOpts.Token, *url)
}
