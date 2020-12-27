package github

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
	"github.com/google/go-github/v32/github"
)

var (
	fc = 0 //frame counter
)

// collectToken gets a personal access token for the user. If there is one saved
// in the config then it'll use that and ask the user if that's the token they
// want to use. If there isn't a token saved, it goes through the steps of
// authenticating with our GitHub OAuth app via the device flow
// https://docs.github.com/en/free-pro-team@latest/developers/apps/authorizing-oauth-apps#device-flow
func collectToken() {
	defer func() { // make sure the client gets placed with the authenticated one
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
		printCode(query.UserCode, query.ExpiresIn)
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
	cache.token = pollResults.AccessToken
}

// pollForToken polls GitHub for the user's PAT. If I poll faster than
// query.Interval, GitHub will rate limit me. As of writing, the interval is
// 5 seconds, and abusing the rate limit adds 5 seconds.
func pollForToken(query *types.GHOAuthAppQuery) *types.OAuthAppPoll {
	pollReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	pollReq.Header.Add("Accept", "application/json")
	params := pollReq.URL.Query()
	params.Add("client_id", cache.clientID)
	params.Add("device_code", query.DeviceCode)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	pollReq.URL.RawQuery = params.Encode()
	var err error = nil
	pollResp := &types.OAuthAppPoll{}
	interval := query.Interval
	for i := 0; i == 0 || pollResp.AccessToken == "" &&
		pollResp.Error == "authorization_pending"; i++ {
		if pollResp.Error == "slow_down" {
			interval += 5 //to avoid further rate limiting
		}
		// Must wait a certain amount of time or else the API will rate limit me
		wait(interval)
		_, err = client.Do(cache.ctx, pollReq, pollResp)
		utils.CheckError(err)
	}
	fmt.Print("\r")
	cache.token = pollResp.AccessToken
	//now with the token we can use a real authenticated client
	return pollResp
}

// wait prints a pretty little animation while AIT waits for the user's to
// authenticate the app on GitHub.
func wait(seconds int) {
	ticker := time.Tick(time.Second)
	for i := 0; i < seconds; i++ {
		fmt.Printf("\r[%v] Checking in %v second(s)...",
			utils.Spinner[fc%len(utils.Spinner)], seconds-i)
		<-ticker
		fc++
	}
}

// disambiguateError given an errMsg from GitHub, returns a more thorough
// explanation of the error and whether or not it is recoverable.
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

// printCode prints the user's code in a pretty format.
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

// SaveToken saves the user's PAT to the global config and writes the file.
func SaveToken() {
	config.Global.Git.PAT = cache.token
	config.GenConf(config.Global)
}
