package manifest

import (
	"errors"
	"net/url"

	"github.com/arken/ark/manifest/upstream"
	"github.com/go-git/go-git/v5/config"
)

func (m *Manifest) Fork() error {
	url, err := url.Parse(m.url)
	if err != nil {
		return err
	}

	// Check for matching upstream.
	upstream, ok := upstream.AvailableUpstreams[url.Host]
	if !ok {
		return errors.New("unknown upstream")
	}

	// If an upstream if found use it to fork the repository.
	m.forkUrl, err = upstream.Fork(m.gitOpts.Token, *url)
	if err != nil {
		return err
	}

	_, err = m.r.CreateRemote(&config.RemoteConfig{
		Name: "fork",
		URLs: []string{m.forkUrl},
	})
	if err != nil && err.Error() == "remote already exists" {
		return nil
	}
	return err
}
