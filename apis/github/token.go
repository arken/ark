package github

import (
	"fmt"
	"golang.org/x/oauth2"
	"math"
	"net/http"
	"time"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"

	"github.com/google/go-github/v32/github"
)

func collectToken() {
	defer func() {
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cache.token},
		)
		client = github.NewClient(oauth2.NewClient(cache.ctx, tokenSource))
	}()
	if cache.token != "" {
		return
	}
	if cache.clientID == "" {
		utils.FatalPrintln("Need a client ID in the environment if no token is provided!")
	}
	req, _ := http.NewRequest("POST", "https://github.com/login/device/code", nil)
	req.Header.Add("Accept", "application/json")
	params := req.URL.Query()
	params.Add("client_id", cache.clientID)
	params.Add("scope", "repo")
	req.URL.RawQuery = params.Encode()
	var pollResults *types.OAuthAppPoll
	for {
		query := &types.GHOAuthAppQuery{}
		_, err := client.Do(cache.ctx, req, query)
		if err != nil {
			utils.FatalPrintln(`Something went wrong while trying to contact GitHub.
Is this computer connected to the internet?`)
		}
		printCode(query.User_code, query.Expires_in)
		pollResults = pollForToken(query)
		if pollResults.Error == "authorization_pending" {
			break
		}
		msg, fatal := disambiguateError(pollResults.Error)
		if fatal {
			utils.FatalPrintln(msg)
		}
		fmt.Println(msg)
	}
	cache.token = pollResults.Access_token
}

func pollForToken(query *types.GHOAuthAppQuery) *types.OAuthAppPoll {
	pollReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	pollReq.Header.Add("Accept", "application/json")
	params := pollReq.URL.Query()
	params.Add("client_id", cache.clientID)
	params.Add("device_code", query.Device_code)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	pollReq.URL.RawQuery = params.Encode()
	var err error = nil
	pollResp := &types.OAuthAppPoll{}
	interval := query.Interval
	for i := 0; i == 0 || pollResp.Access_token == "" &&
		pollResp.Error == "authorization_pending"; i++ {
		if pollResp.Error == "slow_down" {
			interval += 5 //to avoid further rate limiting
		}
		// Must wait a certain amount of time or else the API will rate limit me
		time.Sleep(time.Duration(interval) * time.Second)
		_, err = client.Do(cache.ctx, pollReq, pollResp)
		utils.CheckError(err)
	}
	cache.token = pollResp.Access_token
	//now with the token we can use a real authenticated client
	return pollResp
}

func disambiguateError(errMsg string) (string, bool) {
	var msg string
	fatal := true
	switch errMsg {
	case "expired_token":
		msg = "You didn't authorize the app in time! Please try again."
		fatal = false
		break
	case "unsupported_grant_type":
		msg = "Unsupported grant type! Please let the AIT devs know you got this error."
		break
	case "incorrect_device_code":
		msg = "Wrong device code! Please let the AIT devs know you got this error."
		break
	case "access_denied":
		msg = "You didn't give the application access to the account! You must do this in order to submit."
	case "incorrect_client_credentials":
		msg = "Incorrect client credentials! Please let the AIT devs know you got this error."
		break
	default:
		msg = fmt.Sprintf("Unexpected error: %v. Please let the AIT devs know about this.", errMsg)
	}
	return msg, fatal
}

func printCode(code string, expiry int) {
	now := time.Now()
	expireTime := now.Add(time.Duration(expiry) * time.Second)
	minutes := math.Round(float64(expiry) / 60.0)
	fmt.Printf(
`Go to https://github.com/login/device and enter the following code. You should
see a request to authorize "AIT GitHub Worker". Please authorize this request, but 
not if it's from anyone other than AIT GitHub Worker by arkenproject!
=================================================================
                            %v
=================================================================
This code will expire in about %v minutes at %v.

`, code, int(minutes), expireTime.Format("3:04 PM"))
}

func SaveToken() {
	config.Global.Git.PAT = cache.token
	config.GenConf(config.Global)
}
