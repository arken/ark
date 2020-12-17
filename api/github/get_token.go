package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
)



func GetToken() {
	if GHInfo.token != "" {
		return
	}
	if GHInfo.clientID == "" {
		utils.FatalPrintln("Need a client ID in the environment!")
	}
	req, _ := http.NewRequest("POST", "https://github.com/login/device/code", nil)
	req.Header.Add("Accept", "application/json")
	params := req.URL.Query()
	params.Add("client_id", GHInfo.clientID)
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
		fmt.Println("Go to https://github.com/login/device and enter the following device code")
		fmt.Printf(`=================================================================
                            %s
=================================================================
You have %v minutes to do enter the code.
`, query.User_code, query.Expires_in / 60)
		pollResults, err = pollForToken(query, client, 30) //30 second time out
		if err != nil {
			msg, fatal := disambiguateError(err)
			if fatal {
				utils.FatalPrintln(msg)
			}
			fmt.Println(msg)
		} else {
			break
		}
	}
	GHInfo.token = pollResults.Access_token
	greet()
}

func greet() {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GHInfo.token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))
	auth, _, err := client.Authorizations.Check(ctx, GHInfo.clientID, GHInfo.token)
	utils.CheckError(err)
	fmt.Printf("Authenticated as %v\n", auth.User.Name)
	GHInfo.User = auth.User
}

func pollForToken(query *types.GHOAuthAppQuery, client http.Client, timeout int) (*types.OAuthAppPoll, error) {
	pollReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	pollReq.Header.Add("Accept", "application/json")
	params := pollReq.URL.Query()
	params.Add("client_id", GHInfo.clientID)
	params.Add("device_code", query.Device_code)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	pollReq.URL.RawQuery = params.Encode()
	var resp *http.Response
	var err error = nil
	pollResp := &types.OAuthAppPoll{}
	interval := query.Interval
	elapsed := 0
	for i := 0; i == 0 || pollResp.Error == "authorization_pending"; i++ {
		if elapsed >= timeout {
			return nil, fmt.Errorf("timed out")
		} else if pollResp.Error == "slow_down" {
			interval += 5 //to avoid further rate limiting
		}
		// Must wait a certain amount of time or else the API will rate limit me
		time.Sleep(time.Duration(interval) * time.Second)
		resp, err = client.Do(pollReq)
		if resp != nil {
			scanJsonToStruct(resp.Body, pollResp)
		}
		elapsed += interval
	}
	GHInfo.token = pollResp.Access_token
	return pollResp, err
}

func disambiguateError(err error) (string, bool) {
	var msg string
	fatal := true
	switch err.Error() {
	case "expired_token":
		msg = "You didn't authorize the app in time! Please try again."
		fatal = false
		break
	case "unsupported_grant_type":
		msg = "Unsupported grant type! Please let the AIT devs know you got this error."
		break
	case "incorrect_device_code":
		msg = "You mistyped the device code! Please try again."
		fatal = false
		break
	case "access_denied":
		msg = "You didn't give the application access to the account! You must do this in order to submit."
	case "incorrect_client_credentials":
		msg = "Incorrect client credentials! Please let the AIT devs know you got this error."
		break
	default:
		msg = fmt.Sprintf("Unexpected error: %v. Please let the AIT devs know about this.",err.Error())
		fatal = true
	}
	return msg, fatal
}

func scanJsonToStruct(jData io.Reader, toFill interface{}) {
	decoder := json.NewDecoder(jData)
	err := decoder.Decode(toFill)
	if err != nil {
		utils.FatalPrintln("Received malformed JSON:", err)
	}
}
