package manifest

import (
	"errors"
	"net/url"

	"github.com/arken/ark/manifest/upstream"
)

func (m *Manifest) OpenPR(mainBranch, prTitle, prBody string) error {
	url, err := url.Parse(m.url)
	if err != nil {
		return err
	}

	// Check for matching upstream.
	up, ok := upstream.AvailableUpstreams[url.Host]
	if !ok {
		return errors.New("unknown upstream")
	}

	fork, err := url.Parse(m.forkUrl)
	if err != nil {
		return err
	}

	prBranch, err := m.GetBranchName()
	if err != nil {
		return err
	}

	opts := upstream.PrOpts{
		Origin:     *url,
		Fork:       *fork,
		Token:      m.gitOpts.Token,
		MainBranch: mainBranch,
		PrBranch:   prBranch,
		PrTitle:    prTitle,
		PrBody:     prBody,
	}

	// Attempt to open a PR using the found upstream.
	return up.OpenPR(opts)
}

func (m *Manifest) SearchPrByBranch(branchName string) error {
	url, err := url.Parse(m.url)
	if err != nil {
		return err
	}

	// Check for matching upstream.
	upstream, ok := upstream.AvailableUpstreams[url.Host]
	if !ok {
		return errors.New("unknown upstream")
	}

	return upstream.SearchPrByBranch(*url, m.gitOpts.Token, branchName)
}
