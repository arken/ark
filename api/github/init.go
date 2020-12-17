package github

import (
	"context"
	"github.com/arkenproject/ait/utils"
	"golang.org/x/oauth2"
	"os"

	"github.com/arkenproject/ait/config"

	"github.com/google/go-github/v32/github"
)

type Info struct {
	user      *github.User
	repo      Repository
	token     string
	clientID  string
	keysetSHA string
}

type Repository struct {
	url   string
	owner string
	name  string
}

var (
	cache = Info{}
)

func init() {
	cache.clientID = os.Getenv("GHA_CLIENT_ID")
	cache.token = config.Global.Git.PAT
}

func getClient() *github.Client {
	if cache.token == "" {
		panic("No token yet!")
	}
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cache.token},
	)
	return github.NewClient(oauth2.NewClient(context.Background(), tokenSource))
}

func SetURL(URL string) {
	cache.repo = Repository{
		url:   URL,
		owner: utils.GetRepoOwner(URL),
		name:  utils.GetRepoName(URL),
	}
}
