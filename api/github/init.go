package github

import (
	"context"
	"golang.org/x/oauth2"
	"os"

	"github.com/arkenproject/ait/config"

	"github.com/google/go-github/v32/github"
)

type Info struct {
	User *github.User
	Repo Repository
	token string
	OverwriteIfPresent bool
	clientID string
}

type Repository struct {
	URL string
	Owner string
	Name string
}

var (
	Cache = Info{}
)

func init() {
	Cache.clientID = os.Getenv("GHA_CLIENT_ID")
	Cache.token = config.Global.Git.PAT
}

func getClient() *github.Client {
	if Cache.token == "" {
		panic("No token yet!")
	}
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: Cache.token},
	)
	return github.NewClient(oauth2.NewClient(context.Background(), tokenSource))
}
