package types

type GHOAuthAppQuery struct {
	Device_code      string
	User_code        string
	Verification_uri string
	Expires_in       int
	Interval         int
}

type OAuthAppPoll struct {
	Access_token string
	Token_type   string
	Scope        string
	Error        string
}

