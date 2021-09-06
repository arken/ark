package manifest

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Push performs a "git push" on the repository.
func (m *Manifest) Push() (err error) {
	h, err := m.r.Head()
	if err != nil {
		return err
	}
	// Generate <src>:<dest> reference string
	refStr := h.Name().String() + ":" + h.Name().String()
	// Push Branch to Origin
	err = m.r.Push(&git.PushOptions{
		RemoteName: "fork",
		RefSpecs:   []config.RefSpec{config.RefSpec(refStr)},
		Auth: &http.BasicAuth{
			Username: m.gitOpts.Username,
			Password: m.gitOpts.Token,
		},
	})
	return err
}
