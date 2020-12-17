package api

import (
	"encoding/json"
	"fmt"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
	"io"
	"net/http"
	"os"
	"time"
)

func GetToken() string {
	req, _ := http.NewRequest("POST", "https://github.com/login/device/code", nil)
	req.Header.Add("Accept", "application/json")
	params := req.URL.Query()
	params.Add("client_id", os.Getenv("client_id")) //temporary
	params.Add("scope", os.Getenv("repo"))
	req.URL.RawQuery = params.Encode()
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.FatalPrintln(`Something went wrong while trying to contact GitHub.
Is this computer connected to the internet?`)
	}
	first := &types.GHOAuthAppQuery{}
	scanJsonToStruct(resp.Body, first)
	fmt.Println("Go to https://github.com/login/device and enter the following code")
	fmt.Printf(`=================================================================
                            %s
=================================================================
`, first.User_code)
	pollResults, err := pollForToken(first, client, 30)
	if err != nil {
		disambiguateError(err)
	}
	return pollResults.Access_token
}

func disambiguateError(err error) {
	
}

func pollForToken(query *types.GHOAuthAppQuery, client http.Client, timeout int) (*types.OAuthAppPoll, error) {
	pollReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	pollReq.Header.Add("Accept", "application/json")
	params := pollReq.URL.Query()
	params.Add("client_id", "c3804a6f9da1211bcc93")
	params.Add("device_code", query.Device_code)
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	pollReq.URL.RawQuery = params.Encode()
	var resp *http.Response
	var err error = nil
	pollResp := &types.OAuthAppPoll{}
	for i := 0; i == 0 || pollResp.Error == "authorization_pending"; i++ {
		if i * query.Interval > timeout {
			return nil, fmt.Errorf("timed out")
		}
		time.Sleep(time.Duration(query.Interval) * time.Second)
		resp, err = client.Do(pollReq)
		if resp != nil {
			scanJsonToStruct(resp.Body, pollResp)
		}
	}
	return pollResp, err
}

func scanJsonToStruct(jData io.Reader, toFill interface{}) {
	decoder := json.NewDecoder(jData)
	err := decoder.Decode(toFill)
	if err != nil {
		utils.FatalPrintln("Received malformed JSON:", err)
	}
}
