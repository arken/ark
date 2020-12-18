package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"

	"github.com/google/go-github/v32/github"
)

func getToken() {
	if cache.token != "" {
		return
	}
	if cache.clientID == "" {
		utils.FatalPrintln("Need a client ID in the environment!")
	}
	req, _ := http.NewRequest("POST", "https://github.com/login/device/code", nil)
	req.Header.Add("Accept", "application/json")
	params := req.URL.Query()
	params.Add("client_id", cache.clientID)
	params.Add("scope", os.Getenv("repo"))
	req.URL.RawQuery = params.Encode()
	client := http.Client{}
	var pollResults *types.OAuthAppPoll
	for {
		resp, err := client.Do(req)
		if err != nil {
			utils.FatalPrintln(`Something went wrong while trying to contact GitHub.
Is this computer connected to the internet?`)
		}
		query := &types.GHOAuthAppQuery{}
		scanJsonToStruct(resp.Body, query)
		expireTime := time.Now().Add(time.Duration(query.Expires_in) * time.Second)
		fmt.Println("Go to https://github.com/login/device and enter the following code")
		fmt.Printf(`=================================================================
                            %s
=================================================================
This code will expire at %v.
`, query.User_code, expireTime.Format(time.RFC822))
		pollResults = pollForToken(query, client)
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
	greet()
}

func pollForToken(query *types.GHOAuthAppQuery, client http.Client) *types.OAuthAppPoll {
	pollReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	pollReq.Header.Add("Accept", "application/json")
	params := pollReq.URL.Query()
	params.Add("client_id", cache.clientID)
	params.Add("device_code", query.Device_code)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	pollReq.URL.RawQuery = params.Encode()
	var resp *http.Response
	var err error = nil
	pollResp := &types.OAuthAppPoll{}
	interval := query.Interval
	elapsed := 0
	for i := 0; i == 0 || pollResp.Access_token == "" &&
		pollResp.Error == "authorization_pending"; i++ {
		if pollResp.Error == "slow_down" {
			interval += 5 //to avoid further rate limiting
		}
		// Must wait a certain amount of time or else the API will rate limit me
		time.Sleep(time.Duration(interval) * time.Second)
		resp, err = client.Do(pollReq)
		utils.CheckError(err)
		if resp != nil {
			scanJsonToStruct(resp.Body, pollResp)
		}
		elapsed += interval
	}
	cache.token = pollResp.Access_token
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
	case "timed_out":
		msg = ""
	default:
		msg = fmt.Sprintf("Unexpected error: %v. Please let the AIT devs know about this.", errMsg)
	}
	return msg, fatal
}

func greet() {
	client := getClient()
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	user := &github.User{}
	_, err := client.Do(cache.ctx, req, user)
	if err != nil {
		utils.FatalPrintln("Unable to authenticate user:", err)
	}
	fmt.Println("Authenticated as user", *user.Login)
	cache.user = user
}

func scanJsonToStruct(jData io.Reader, toFill interface{}) {
	decoder := json.NewDecoder(jData)
	err := decoder.Decode(toFill)
	if err != nil {
		utils.FatalPrintln("Received malformed JSON:", err)
	}
}

func SaveToken() {
	config.Global.Git.PAT = cache.token
	config.GenConf(config.Global)
}
