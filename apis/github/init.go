package github

import (
	"context"
	"os"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/utils"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type Info struct {
	user      *github.User
	fork      *Repository
	upstream  *Repository
	token     string
	clientID  string
	keysetSHA string
	isPR 	  bool
	ctx 	  context.Context
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
	cache.ctx = context.Background()
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

func Init(URL string, isPR bool) {
	cache.upstream = &Repository{
		url:   URL,
		owner: utils.GetRepoOwner(URL),
		name:  utils.GetRepoName(URL),
	}
	cache.isPR = isPR
	getToken()
}
