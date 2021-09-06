package upstream

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/google/go-github/v38/github"
	"golang.org/x/oauth2"
)

var (
	GitHubClientID string
)

// GitHub is a wrapper struct for the GitHub Upstream
type GitHub struct {
}

type GitHubAppAuthQuery struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func init() {
	registerUpstream(&GitHub{}, "github.com")
}

func (g *GitHub) Auth(path string) (result Guard, err error) {
	// Check if GitHubClientID has not been set.
	if GitHubClientID == "" {
		return nil, errors.New("required client id is nil")
	}

	client := github.NewClient(nil)
	ctx := context.Background()
	defer ctx.Done()
	query := &GitHubAppAuthQuery{}

	// Construct GitHub HTTP query
	req, _ := http.NewRequest("POST", "https://github.com/login/device/code", nil)
	req.Header.Add("Accept", "application/json")

	// Add parameters to request query
	params := req.URL.Query()
	params.Add("client_id", GitHubClientID)
	params.Add("scope", "public_repo")

	// Encode query
	req.URL.RawQuery = params.Encode()

	// Launch request
	_, err = client.Do(ctx, req, query)
	if err != nil {
		return nil, err
	}

	return &GitHubGuard{
			client: client,
			query:  query,
		},
		nil
}

type GitHubGuard struct {
	client *github.Client
	query  *GitHubAppAuthQuery
	token  string
}

type GitHubAppAuthPoll struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

func (g *GitHubGuard) GetAccessToken() string {
	return g.token
}

func (g *GitHubGuard) GetCode() string {
	return g.query.UserCode
}

func (g *GitHubGuard) GetExpireInterval() int {
	return g.query.ExpiresIn
}

func (g *GitHubGuard) GetInterval() int {
	return g.query.Interval
}

func (g *GitHubGuard) CheckStatus() (status string, err error) {
	ctx := context.Background()
	defer ctx.Done()

	// Construct poll request.
	pollReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	pollReq.Header.Add("Accept", "application/json")

	// Add parameters to poll request.
	params := pollReq.URL.Query()
	params.Add("client_id", GitHubClientID)
	params.Add("device_code", g.query.DeviceCode)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	// Construct poll response.
	pollResp := &GitHubAppAuthPoll{}

	// Encode request
	pollReq.URL.RawQuery = params.Encode()

	// Launch request
	_, err = g.client.Do(ctx, pollReq, pollResp)
	if err != nil {
		return "", err
	}

	// Set token on successful auth.
	if pollResp.AccessToken != "" {
		g.token = pollResp.AccessToken
	}

	return pollResp.Error, nil
}

func (g *GitHubGuard) GetUser() (string, error) {
	ctx := context.Background()
	defer ctx.Done()

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))

	// Construct user request to get the name of the current logged in user.
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	user := &github.User{}

	// Launch request
	_, err := client.Do(ctx, req, user)
	if err != nil {
		return "", err
	}

	return *user.Login, nil
}

func (g *GitHub) HaveWriteAccess(token string, url url.URL) (bool, error) {
	// Setup GitHub client
	ctx := context.Background()
	defer ctx.Done()

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))

	username, err := getUsername(client)
	if err != nil {
		return false, err
	}

	// Check a user's given permission level for a repository on GitHub.
	perm, resp, err := client.Repositories.GetPermissionLevel(
		ctx,
		filepath.Base(filepath.Dir(url.Path)),
		filepath.Base(url.Path),
		username,
	)
	if resp != nil && resp.Response.StatusCode != 200 {
		return false, err
	}
	return *perm.Permission == "admin" || *perm.Permission == "write", nil
}

func (g *GitHub) Fork(token string, url url.URL) (string, error) {
	// Setup GitHub client
	ctx := context.Background()
	defer ctx.Done()

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))

	username, err := getUsername(client)
	if err != nil {
		return "", err
	}

	// Check for the existence of the fork before attempting to create one
	remoteRepo, _, err := client.Repositories.Get(
		ctx,
		username,
		filepath.Base(url.Path),
	)
	if err != nil {
		remoteRepo, response, err := client.Repositories.CreateFork(
			ctx,
			filepath.Base(filepath.Dir(url.Path)),
			filepath.Base(url.Path),
			nil,
		)
		if remoteRepo == nil || response.StatusCode != 202 && response.StatusCode != 200 {
			return "", err
		}
	}
	return remoteRepo.GetHTMLURL(), nil
}

func getUsername(client *github.Client) (string, error) {
	// Setup GitHub client
	ctx := context.Background()
	defer ctx.Done()

	// Construct user request to get the name of the current logged in user.
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	user := &github.User{}

	// Launch request
	_, err := client.Do(ctx, req, user)
	if err != nil {
		return "", err
	}
	return *user.Login, nil
}

// OpenPR opens a pull request from the input branch to the destination branch.
func (g *GitHub) OpenPR(opts PrOpts) (err error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	forkOwner := filepath.Base(filepath.Dir(opts.Fork.Path))

	pr := &github.NewPullRequest{
		Title:               github.String(opts.PrTitle),
		Body:                github.String(opts.PrBody),
		Head:                github.String(fmt.Sprintf("%s:%s", forkOwner, opts.PrBranch)),
		Base:                github.String(opts.MainBranch),
		MaintainerCanModify: github.Bool(true),
	}

	repoOwner := filepath.Base(filepath.Dir(opts.Origin.Path))
	repoName := filepath.Base(opts.Origin.Path)

	_, _, err = client.PullRequests.Create(ctx, repoOwner, repoName, pr)
	return err
}

// SearchPrByBranch checks to see if there is an existing PR based on a specific branch
// and if so returns the name.
func (g *GitHub) SearchPrByBranch(url url.URL, token, branchName string) (err error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	result, _, err := client.Search.Issues(
		ctx,
		fmt.Sprintf(
			"head:%s type:pr repo:%s/%s",
			branchName,
			filepath.Dir(url.Path),
			filepath.Base(url.Path),
		), &github.SearchOptions{})
	if err != nil {
		return err
	}
	if len(result.Issues) > 0 {
		for _, issue := range result.Issues {
			if issue.GetState() == "open" {
				return nil
			}
		}
	}
	return errors.New("not found")
}
