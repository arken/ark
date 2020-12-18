package github

import (
	"context"
	"net/http"
	"os"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/utils"

	"github.com/google/go-github/v32/github"
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
	client *github.Client
)

func Init(URL string, isPR bool) {
	client = github.NewClient(&http.Client{}) //basic client for setting up app
	cache.clientID = os.Getenv("GHA_CLIENT_ID")
	cache.token = config.Global.Git.PAT
	cache.ctx = context.Background()
	cache.upstream = &Repository{
		url:   URL,
		owner: utils.GetRepoOwner(URL),
		name:  utils.GetRepoName(URL),
	}
	cache.isPR = isPR
	collectToken()
}
