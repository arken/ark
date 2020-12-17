package github

import (
	"os"

	"github.com/arkenproject/ait/config"

	"github.com/google/go-github/v32/github"
)

type Info struct {
	User *github.User
	URL string
	token string
	OverwriteIfPresent bool
	clientID string
}

var (
	GHInfo = Info{}
)

func init() {
	GHInfo.clientID = os.Getenv("GHA_CLIENT_ID")
	GHInfo.token = config.Global.Git.PAT
}
