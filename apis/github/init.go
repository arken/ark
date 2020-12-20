package github

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/utils"

	"github.com/google/go-github/v32/github"
)

type Info struct {
	user     *github.User
	fork     *Repository
	upstream *Repository
	token    string
	clientID string
	shas	 map[string]string
	isPR     bool
	ctx      context.Context
}

type Repository struct {
	url   string
	owner string
	name  string
}

var (
	cache Info
	client *github.Client
)

// Init sets up the github portion of AIT with the context it needs going
// forward, including the url and client id.
func Init(URL string, isPR bool) bool {
	cache = Info{
		upstream: &Repository{
			url:   URL,
			owner: utils.GetRepoOwner(URL),
			name:  utils.GetRepoName(URL),
		},
		token:    config.Global.Git.PAT,
		clientID: os.Getenv("GHA_CLIENT_ID"),
		shas:     make(map[string]string),
		isPR:     isPR,
		ctx:      context.Background(),
	}
	for correctUser := false; !correctUser; {
		client = github.NewClient(&http.Client{}) //basic client for setting up app
		collectToken()
		correctUser = promptIsCorrectUser()
	}
	if isPR {
		return true
	}
	return hasWritePermission()
}

// promptIsCorrectUser asks the user if the user we authenticated is correct.
// This is necessary for if a user chooses to save their token, but then comes
// back and wants to be a different user. Also, if someone else is already
// logged into GitHub on the user's browser, the login page doesn't give users
// a chance to log in as the account they want and some users may just click
// through without realizing it's not their GH account.
func promptIsCorrectUser() bool {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	user := &github.User{}
	_, err := client.Do(cache.ctx, req, user)
	if err != nil {
		utils.FatalPrintln("Unable to authenticate user!")
	}
	fmt.Println("Successfully authenticated as user", *user.Login)
	cache.user = user
	fmt.Printf("Is this correct? ([y]/n) ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "n" {
		fmt.Println("NNNN")
		cache.token = ""
		SaveToken() //clear the token from config
		return false
	}
	return true
}
